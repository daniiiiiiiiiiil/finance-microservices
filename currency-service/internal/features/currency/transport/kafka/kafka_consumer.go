package kafka

import (
	"encoding/json"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/currency-service/internal/core/kafka"
	service_currency "github.com/daniiiiiiiiiiil/finance-microservices/currency-service/internal/features/currency/service"
	"github.com/daniiiiiiiiiiil/finance-microservices/currency-service/pkg/logger"
	"golang.org/x/net/context"
)

type CurrencyKafkaConsumer struct {
	consumer *kafka.Consumer
	service  *service_currency.CurrencyService
	logger   *logger.Logger
}

func NewCurrencyKafkaConsumer(
	consumer *kafka.Consumer,
	service *service_currency.CurrencyService,
	logger *logger.Logger,
) *CurrencyKafkaConsumer {
	return &CurrencyKafkaConsumer{
		consumer: consumer,
		service:  service,
		logger:   logger,
	}
}

func (c *CurrencyKafkaConsumer) Start(ctx context.Context) error {
	c.consumer.RegisterHandler(kafka.EventTypeTransactionCreated, c.handleTransactionCreated)
	c.consumer.RegisterHandler(kafka.EventTypeTransactionDeleted, c.handleTransactionDeleted)

	return c.consumer.Start(ctx)
}

func (c *CurrencyKafkaConsumer) handleTransactionCreated(ctx context.Context, event kafka.Event) error {
	var currencyEvent kafka.TransactionCurrencyEvent
	if err := json.Unmarshal(event.Data, &currencyEvent); err != nil {
		return fmt.Errorf("failed to unmarshal transaction event: %w", err)
	}

	return c.service.ConvertTransactionToUSD(ctx, currencyEvent)
}

func (c *CurrencyKafkaConsumer) handleTransactionDeleted(ctx context.Context, event kafka.Event) error {
	var currencyEvent kafka.TransactionCurrencyEvent
	if err := json.Unmarshal(event.Data, &currencyEvent); err != nil {
		return fmt.Errorf("failed to unmarshal transaction event: %w", err)
	}

	return c.service.DeleteConvertedUSD(ctx, currencyEvent.TransactionID)
}
