package service_currency

import (
	"backend/internal/core/cache"
	"backend/internal/core/kafka"
	"backend/internal/features/currency/repository/redis"
	"backend/pkg/logger"
	"net/http"
	"time"
)

var _ CurrencyServiceInterface = (*CurrencyService)(nil)

type CurrencyClient struct {
	client  *http.Client
	baseURL string
	logger  logger.Logger
}

func NewCurrencyClient(baseURL string, logger logger.Logger) *CurrencyClient {
	return &CurrencyClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: baseURL,
		logger:  logger,
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
