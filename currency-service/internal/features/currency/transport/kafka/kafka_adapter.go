package kafka

import (
	"context"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/currency-service/internal/core/kafka"
)

type CurrencyEventPublisher struct {
	producer *kafka.Producer
}

func NewCurrencyEventPublisher(producer *kafka.Producer) *CurrencyEventPublisher {
	return &CurrencyEventPublisher{producer: producer}
}

func (p *CurrencyEventPublisher) Publish(ctx context.Context, eventType string, data interface{}) error {
	if err := p.producer.SendEvent(ctx, eventType, data); err != nil {
		return fmt.Errorf("publish event: %w", err)
	}
	return nil
}

type CurrencyEventConsumer struct {
	consumer *kafka.Consumer
}

func NewCurrencyEventConsumer(consumer *kafka.Consumer) *CurrencyEventConsumer {
	return &CurrencyEventConsumer{consumer: consumer}
}

func (c *CurrencyEventConsumer) Start(ctx context.Context) error {
	return c.consumer.Start(ctx)
}

func (c *CurrencyEventConsumer) RegisterHandler(eventType string, handler func(ctx context.Context, event kafka.Event) error) {
	c.consumer.RegisterHandler(eventType, handler)
}

func (c *CurrencyEventConsumer) Close() error {
	return c.consumer.Close()
}
