package kafka

import (
	"context"

	corekafka "github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/core/kafka"
)

type Handler func(ctx context.Context, event corekafka.Event) error

type Consumer struct {
	consumer *corekafka.Consumer
}

func NewConsumer(c *corekafka.Consumer) *Consumer {
	return &Consumer{consumer: c}
}

func (c *Consumer) RegisterHandler(eventType string, h Handler) error {
	c.consumer.RegisterHandler(eventType, h)
	return nil
}

func (c *Consumer) Start(ctx context.Context) error {
	return c.consumer.Start(ctx)
}

func (c *Consumer) Close() error {
	return c.consumer.Close()
}
