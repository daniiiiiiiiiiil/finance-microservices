package kafka

import (
	"context"

	corekafka "github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/core/kafka"
	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/pkg/logger"
	"go.uber.org/zap"
)

type EventHandler func(ctx context.Context, event corekafka.Event) error

type AuthKafkaConsumer struct {
	consumer *corekafka.Consumer
	handlers map[string]EventHandler
	logger   *logger.Logger
}

func NewAuthKafkaConsumer(
	consumer *corekafka.Consumer,
	logger *logger.Logger,
) *AuthKafkaConsumer {
	return &AuthKafkaConsumer{
		consumer: consumer,
		handlers: make(map[string]EventHandler),
		logger:   logger,
	}
}

func (c *AuthKafkaConsumer) RegisterHandler(eventType string, handler EventHandler) error {
	c.handlers[eventType] = handler

	c.consumer.RegisterHandler(eventType, handler)

	c.logger.Debug("registered handler", zap.String("event_type", eventType))
	return nil
}

func (c *AuthKafkaConsumer) Start(ctx context.Context) error {
	c.logger.Info("starting auth kafka consumer")
	return c.consumer.Start(ctx)
}

func (c *AuthKafkaConsumer) Close() error {
	c.logger.Info("closing auth kafka consumer")
	return c.consumer.Close()
}
