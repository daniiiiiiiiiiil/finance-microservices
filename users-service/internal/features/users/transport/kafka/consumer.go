package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	corekafka "github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/kafka"
	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/pkg/logger"
	"go.uber.org/zap"
)

type EventHandler func(ctx context.Context, event ports.Event) error

type UserKafkaConsumer struct {
	consumer *corekafka.Consumer
	handlers map[string]EventHandler
	logger   *logger.Logger
}

func NewUserKafkaConsumer(
	consumer *corekafka.Consumer,
	logger *logger.Logger,
) *UserKafkaConsumer {
	return &UserKafkaConsumer{
		consumer: consumer,
		handlers: make(map[string]EventHandler),
		logger:   logger,
	}
}

func (c *UserKafkaConsumer) RegisterHandler(eventType string, handler EventHandler) error {
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

func (c *UserKafkaConsumer) Start(ctx context.Context) error {
	c.logger.Info("starting user kafka consumer")
	return c.consumer.Start(ctx)
}

func (c *UserKafkaConsumer) Close() error {
	c.logger.Info("closing user kafka consumer")
	return c.consumer.Close()
}

func ParseUserEvent(data []byte) (map[string]interface{}, error) {
	var event map[string]interface{}
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("unmarshal user event: %w", err)
	}
	return event, nil
}
