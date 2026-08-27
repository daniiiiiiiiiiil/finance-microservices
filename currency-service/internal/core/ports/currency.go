package ports

import (
	"context"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/currency-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/currency-service/internal/core/kafka"
)

type CurrencyServiceInterface interface {
	GetRates(ctx context.Context, base string) (*domain.Rate, error)
	Convert(ctx context.Context, convert *domain.Conversion) (*domain.Conversion, error)
	ConvertBatch(ctx context.Context, from string, toList []string, amount float64) (map[string]float64, error)
	GetTransactionUSD(ctx context.Context, txID int) (domain.TransactionUSD, error)
	FetchRates(ctx context.Context, base string) (*domain.Rate, error)
	StartConsumer(ctx context.Context)
}

type RateCacheInterface interface {
	GetRate(ctx context.Context, base string) (*domain.Rate, error)
	SetRate(ctx context.Context, rate domain.Rate, ttl time.Duration) error
	GetConvertedUSD(ctx context.Context, txID int) (float64, error)
	SetConvertedUSD(ctx context.Context, txID int, amount float64, ttl time.Duration) error
	DeleteRate(ctx context.Context, base string) error
	DeleteConvertedUSD(ctx context.Context, txID int) error
	Exists(ctx context.Context, key string) (int64, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	SetTransactionUSD(ctx context.Context, tx domain.TransactionUSD, ttl time.Duration) error
	GetTransactionUSD(ctx context.Context, txID int) (domain.TransactionUSD, error)
}

type CurrencyAPIClientInterface interface {
	GetRatesFromAPI(ctx context.Context, base string) (*domain.Rate, error)
}

type EventPublisherInterface interface {
	Publish(ctx context.Context, eventType string, data interface{}) error
}

type EventConsumerInterface interface {
	Start(ctx context.Context) error
	RegisterHandler(eventType string, handler func(ctx context.Context, event kafka.Event) error)
	Close() error
}
