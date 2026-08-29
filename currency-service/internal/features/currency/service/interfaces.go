package service_currency

import (
	"context"
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
	ConvertTransactionToUSD(ctx context.Context, tx kafka.TransactionCurrencyEvent) error
	DeleteConvertedUSD(ctx context.Context, txID int) error
}

type CurrencyAPIClientInterface interface {
	GetRatesFromAPI(ctx context.Context, base string) (*domain.Rate, error)
}
