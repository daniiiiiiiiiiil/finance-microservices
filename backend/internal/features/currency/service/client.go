package service_currency

import (
	"backend/internal/core/cache"
	"backend/internal/core/domain"
	"backend/internal/core/kafka"
	"backend/internal/core/logger"
	"backend/internal/features/currency/repository/redis"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type CurrencyClient struct {
	client  *http.Client
	baseURL string
}

func NewCurrencyClient(baseURL string) *CurrencyClient {
	return &CurrencyClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: baseURL,
	}
}

type CurrencyService struct {
	rateCache *redis.RateCache
	client    *CurrencyClient
	logger    *logger.Logger
	redis     cache.RedisInterface
	producer  *kafka.Producer
}

func NewCurrencyService(rateCache *redis.RateCache, client *CurrencyClient, logger *logger.Logger, redis cache.RedisInterface, producer *kafka.Producer) *CurrencyService {
	return &CurrencyService{
		rateCache: rateCache,
		client:    client,
		logger:    logger,
		redis:     redis,
		producer:  producer,
	}
}

func (c *CurrencyClient) GetRatesFromAPI(ctx context.Context, base string) (*domain.Rate, error) {
	// Формируем https://api.exchangerate-api.com/v4/latest/USD
	url := fmt.Sprintf("%s/%s", c.baseURL, base)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api returned status: %s", resp.Status)
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
	rate, err := s.rateCache.GetRate(ctx, base)
	if err == nil {
		s.logger.Debug("rates found in cache", zap.String("base", base))

		return rate, nil
	}

	s.logger.Debug("rates not in cache, fetching from API", zap.String("base", base))

	rate, err = s.client.GetRatesFromAPI(ctx, base)
	if err != nil {
		return nil, fmt.Errorf("get rates from API: %w", err)
	}

	if err := s.rateCache.SetRate(ctx, *rate, 1*time.Hour); err != nil {
		s.logger.Warn("failed to cache rates", zap.Error(err))
	}

	return rate, nil
}
