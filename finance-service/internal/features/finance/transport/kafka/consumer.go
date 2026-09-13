package kafka

import (
	"context"

	corekafka "github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/kafka"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/pkg/logger"
	"go.uber.org/zap"
)

type EventHandler func(ctx context.Context, event ports.Event) error

type FinanceKafkaConsumer struct {
	consumer *corekafka.Consumer
	handlers map[string]EventHandler
	logger   *logger.Logger
}

func NewFinanceKafkaConsumer(
	consumer *corekafka.Consumer,
	logger *logger.Logger,
) *FinanceKafkaConsumer {
	return &FinanceKafkaConsumer{
		consumer: consumer,
		handlers: make(map[string]EventHandler),
		logger:   logger,
	}
}

func (c *FinanceKafkaConsumer) RegisterHandler(eventType string, handler EventHandler) error {
	c.handlers[eventType] = handler

	c.consumer.RegisterHandler(eventType, func(ctx context.Context, event corekafka.Event) error {
		portsEvent := ports.Event{
			ID:        event.ID,
			Type:      event.Type,
			Timestamp: event.Timestamp,
			Data:      event.Data,
		}
		return handler(ctx, portsEvent)
	})

	c.logger.Debug("registered handler", zap.String("event_type", eventType))
	return nil
}

func (c *FinanceKafkaConsumer) Start(ctx context.Context) error {
	c.logger.Info("starting finance kafka consumer")
	return c.consumer.Start(ctx)
}

func (c *FinanceKafkaConsumer) Close() error {
	c.logger.Info("closing finance kafka consumer")
	return c.consumer.Close()
}
