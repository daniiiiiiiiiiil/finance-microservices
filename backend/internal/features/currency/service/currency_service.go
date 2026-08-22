package service_currency

import (
	"backend/internal/core/domain"
	"fmt"
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/context"
)

func (s *CurrencyService) GetRates(ctx context.Context, base string) (*domain.Rate, error) {
	s.logger.Debug("GetRates called", zap.String("base", base))

	var rate domain.Rate
	err := s.redis.Get(ctx, fmt.Sprintf("currency:rate:%s", base), &rate)
	if err == nil {
		s.logger.Debug("rates found in Redis", zap.String("base", base))
		return &rate, nil
	}

	s.logger.Debug("rates not in Redis, fetching from API", zap.String("base", base))

	rateFromAPI, err := s.FetchRates(ctx, base)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rates: %w", err)
	}

	if rateFromAPI == nil {
		return nil, fmt.Errorf("rate is nil for base %s", base)
	}

	if err := s.rateCache.SetRate(ctx, *rateFromAPI, time.Hour); err != nil {
		s.logger.Warn("failed to cache rates", zap.Error(err))
	}

	s.logger.Debug("GetRates success", zap.Int("rates_count", len(rateFromAPI.Rates)))
	return rateFromAPI, nil
}

func (s *CurrencyService) Convert(ctx context.Context, convert *domain.Conversion) (*domain.Conversion, error) {
	s.logger.Debug("Convert service called",
		zap.String("from", convert.From),
		zap.String("to", convert.To),
		zap.Float64("amount", convert.Amount),
	)

	if err := convert.Validate(); err != nil {
		s.logger.Error("Validation failed", zap.Error(err))
		return nil, fmt.Errorf("invalid conversion: %w", err)
	}

	s.logger.Debug("Getting rates for", zap.String("base", convert.From))
	rate, err := s.GetRates(ctx, convert.From)
	if err != nil {
		s.logger.Error("GetRates failed", zap.Error(err))
		return nil, fmt.Errorf("failed to get rates: %w", err)
	}

	if rate == nil {
		s.logger.Error("Rate is nil", zap.String("base", convert.From))
		return nil, fmt.Errorf("rates not available for %s", convert.From)
	}

	if rate.Rates == nil {
		s.logger.Error("Rate.Rates is nil", zap.String("base", convert.From))
		return nil, fmt.Errorf("rates map is nil for %s", convert.From)
	}

	s.logger.Debug("Rates received", zap.Any("rates", rate.Rates))

	kursTo, ok := rate.Rates[convert.To]
	if !ok {
		s.logger.Error("Currency not found", zap.String("to", convert.To))
		return nil, fmt.Errorf("currency %s not found", convert.To)
	}

	result := convert.Amount * kursTo
	s.logger.Debug("Conversion result", zap.Float64("result", result))

	return &domain.Conversion{
		From:      convert.From,
		To:        convert.To,
		Amount:    convert.Amount,
		Result:    result,
		Rate:      kursTo,
		Timestamp: time.Now(),
	}, nil
}

func (s *CurrencyService) ConvertBatch(ctx context.Context, from string,
	toList []string,
	amount float64) (map[string]float64, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than zero")
	}
	if len(toList) == 0 {
		return nil, fmt.Errorf("at least one 'to' currency must be specified")
	}
	if len(from) == 0 {
		return nil, fmt.Errorf("'from' currency must not be empty")
	}

	rate, err := s.GetRates(ctx, from)
	if err != nil {
		return nil, fmt.Errorf("failed to get rates: %w", err)
	}

	if rate == nil || rate.Rates == nil {
		return nil, fmt.Errorf("rates not available for %s", from)
	}

	result := make(map[string]float64)
	var missing []string

	for _, to := range toList {
		kursTo, ok := rate.Rates[to]
		if !ok {
			missing = append(missing, to)
			continue
		}
		result[to] = amount * kursTo
	}

	if len(missing) > 0 {
		s.logger.Warn("some currencies not found",
			zap.Strings("missing", missing),
			zap.String("from", from))
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no valid currencies found in %v", toList)
	}

	return result, nil
}

func (s *CurrencyService) GetTransactionUSD(ctx context.Context, txID int) (domain.TransactionUSD, error) {
	if txID <= 0 {
		return domain.TransactionUSD{}, fmt.Errorf("txID must be greater than zero")
	}

	tx, err := s.rateCache.GetTransactionUSD(ctx, txID)
	if err != nil {
		s.logger.Warn("converted USD not found in cache", zap.Int("txID", txID), zap.Error(err))
		return domain.TransactionUSD{}, err
	}

	s.logger.Debug("converted USD found in cache",
		zap.Int("txID", txID),
		zap.Float64("usd", tx.AmountUSD))
	return tx, nil
}
