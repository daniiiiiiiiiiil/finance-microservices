package kafka

import (
	"context"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/currency-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/currency-service/internal/core/kafka"
)

type CurrencyKafkaProducer struct {
	producer *kafka.Producer
}

func NewCurrencyKafkaProducer(producer *kafka.Producer) *CurrencyKafkaProducer {
	return &CurrencyKafkaProducer{producer: producer}
}

func (p *CurrencyKafkaProducer) PublishConvertedEvent(ctx context.Context, tx domain.TransactionUSD) error {
	eventData := kafka.ConvertedEvent{
		TransactionID:    tx.TransactionID,
		AmountUSD:        tx.AmountUSD,
		OriginalAmount:   tx.OriginalAmount,
		OriginalCurrency: tx.OriginalCurrency,
		ConvertedAt:      time.Now(),
	}

	return p.producer.SendEvent(ctx, "currency.converted", eventData)
}
