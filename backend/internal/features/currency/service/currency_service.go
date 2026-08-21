package service_currency

import (
	"backend/internal/core/domain"
	"fmt"
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/context"
)

func (s *CurrencyService) GetRates(ctx context.Context, base string) (*domain.Rate, error) {
	key := fmt.Sprintf("currency:rate:%s", base)

	var rate *domain.Rate
	if err := s.redis.Get(ctx, key, &rate); err != nil {
		return rate, nil

	}

	rateFromAPI, err := s.FetchRates(ctx, base)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rates: %w", err)
	}

	if err := s.rateCache.SetRate(ctx, *rateFromAPI, time.Hour); err != nil {
		s.logger.Warn("failed to cache rates", zap.Error(err))
	}
	return rateFromAPI, nil
}

func (s *CurrencyService) Convert(ctx context.Context, convert *domain.Conversion) (*domain.Conversion, error) {
	if err := convert.Validate(); err != nil {
		return nil, fmt.Errorf("invalid conversion: %w", err)
	}
	key := fmt.Sprintf("currency:convert:%s:%s:%f", convert.From, convert.To, convert.Amount)
	var cached domain.Conversion
	if err := s.redis.Get(ctx, key, &cached); err != nil {
		return &cached, nil
	}

	rate, err := s.GetRates(ctx, convert.From)
	if err != nil {
		return nil, fmt.Errorf("failed to get rates: %w", err)
	}

	kursTo, ok := rate.Rates[convert.To]
	if !ok {
		return nil, fmt.Errorf("rate %s not found in cache", convert.To)
	}
	result := convert.Amount * kursTo

	conversion := &domain.Conversion{
		From:      convert.From,
		To:        convert.To,
		Amount:    convert.Amount,
		Result:    result,
		Rate:      kursTo,
		Timestamp: time.Now(),
	}

	if err := s.redis.Set(ctx, key, conversion, 10*time.Minute); err != nil {
		s.logger.Warn("failed to cache conversion", zap.Error(err))
	}

	return conversion, nil
}

func (s *CurrencyService) ConvertBatch(ctx context.Context, from string,
	toList []string,
	amount float64) (map[string]float64, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than zero")
	}
	if len(toList) == 0 {
		return nil, fmt.Errorf("at least one to to must be specified")
	}
	if len(from) == 0 {
		return nil, fmt.Errorf("from must not be empty")
	}

	key := fmt.Sprintf("currency:convert:%s:%v:%f", from, toList, amount)
	rate, err := s.GetRates(ctx, from)
	if err != nil {
		s.logger.Info("Cache conversion")
	}
	result := make(map[string]float64)
	for _, to := range toList {
		kursTo, ok := rate.Rates[to]
		if !ok {
			return nil, fmt.Errorf("rate %s not found in cache", to)
		}
		result[to] = amount * kursTo
	}
	if err := s.redis.Set(ctx, key, result, 10*time.Minute); err != nil {
		s.logger.Warn("failed to cache batch conversion", zap.Error(err))
	}
	return result, nil
}

func (s *CurrencyService) GetTransactionUSD(ctx context.Context, txID int) (float64, error) {
	if txID <= 0 {
		return 0, fmt.Errorf("txID must be greater than zero")
	}

	usd, err := s.rateCache.GetConvertedUSD(ctx, txID)
	if err != nil {
		s.logger.Warn("converted USD not found in cache", zap.Int("txID", txID), zap.Error(err))
		return 0, err
	}

	s.logger.Debug("converted USD found in cache", zap.Int("txID", txID), zap.Float64("usd", usd))
	return usd, nil
}
