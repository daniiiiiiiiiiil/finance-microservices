package service_currency

import (
	"net/http"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/currency-service/internal/core/cache"
	"github.com/daniiiiiiiiiiil/finance-microservices/currency-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/currency-service/pkg/logger"
)

var _ ports.CurrencyServiceInterface = (*CurrencyService)(nil)

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
	rateCache      ports.RateCacheInterface
	client         ports.CurrencyAPIClientInterface
	logger         *logger.Logger
	redis          cache.RedisInterface
	eventPublisher ports.EventPublisherInterface
	eventConsumer  ports.EventConsumerInterface
}

func NewCurrencyService(
	rateCache ports.RateCacheInterface,
	client ports.CurrencyAPIClientInterface,
	logger *logger.Logger,
	redis cache.RedisInterface,
	eventPublisher ports.EventPublisherInterface,
	eventConsumer ports.EventConsumerInterface,
) *CurrencyService {
	return &CurrencyService{
		rateCache:      rateCache,
		client:         client,
		logger:         logger,
		redis:          redis,
		eventPublisher: eventPublisher,
		eventConsumer:  eventConsumer,
	}
}
