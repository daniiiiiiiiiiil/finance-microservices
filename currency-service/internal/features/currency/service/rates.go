package service_currency

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/currency-service/internal/core/domain"
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

func (c *CurrencyClient) GetRatesFromAPI(ctx context.Context, base string) (*domain.Rate, error) {
	url := fmt.Sprintf("%s?from=%s", c.baseURL, base)
	c.logger.Debug("Calling API", zap.String("url", url))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Warn("API timeout, using fallback rates", zap.String("base", base))
		return getFallbackRates(base), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Warn("API returned error, using fallback rates",
			zap.String("base", base),
			zap.Int("status", resp.StatusCode))
		return getFallbackRates(base), nil
	}

	var apiResponse struct {
		Base  string             `json:"base"`
		Rates map[string]float64 `json:"rates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &domain.Rate{
		Base:      apiResponse.Base,
		Rates:     apiResponse.Rates,
		Timestamp: time.Now(),
	}, nil
}

func (s *CurrencyService) FetchRates(ctx context.Context, base string) (*domain.Rate, error) {
	s.logger.Debug("FetchRates called", zap.String("base", base))
	rate, err := s.rateCache.GetRate(ctx, base)
	if err == nil {
		s.logger.Debug("rates found in cache", zap.String("base", base))

		return rate, nil
	}

	s.logger.Debug("rates not in cache, fetching from API", zap.String("base", base))

	rate, err = s.client.GetRatesFromAPI(ctx, base)
	if err != nil {
		s.logger.Warn("API failed, using fallback rates",
			zap.String("base", base),
			zap.Error(err))
		return nil, fmt.Errorf("get rates from API: %w", err)
	}

	if rate == nil {
		return nil, fmt.Errorf("rate from API is nil for base %s", base)
	}

	if err := s.rateCache.SetRate(ctx, *rate, 1*time.Hour); err != nil {
		s.logger.Warn("failed to cache rates", zap.Error(err))
	}

	return rate, nil
}

func getFallbackRates(base string) *domain.Rate {
	baseRates := map[string]float64{
		"USD": 1.0, "EUR": 1.0, "RUB": 91.5,
		"GBP": 0.78, "JPY": 147.2, "CNY": 7.25,
		"AED": 3.67, "TRY": 48.04,
	}

	baseRate, ok := baseRates[base]
	if !ok {
		baseRate = 1.0
	}

	rates := make(map[string]float64)
	for currency, rate := range baseRates {
		rates[currency] = rate / baseRate
	}

	return &domain.Rate{
		Base:      base,
		Rates:     rates,
		Timestamp: time.Now(),
	}
}
