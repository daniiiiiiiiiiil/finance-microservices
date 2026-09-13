package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/kafka"
	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/pkg/logger"
	"go.uber.org/zap"
)

type SagaKafkaConsumer struct {
	consumer *kafka.Consumer
	handlers map[string]ports.EventHandler
	logger   *logger.Logger
}

func NewSagaKafkaConsumer(
	consumer *kafka.Consumer,
	logger *logger.Logger,
) *SagaKafkaConsumer {
	return &SagaKafkaConsumer{
		consumer: consumer,
		handlers: make(map[string]ports.EventHandler),
		logger:   logger,
	}
}

func (c *SagaKafkaConsumer) RegisterHandler(eventType string, handler ports.EventHandler) error {
	c.handlers[eventType] = handler

	c.consumer.RegisterHandler(eventType, func(ctx context.Context, event kafka.Event) error {
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

func (c *SagaKafkaConsumer) Start(ctx context.Context) error {
	c.logger.Info("starting saga kafka consumer")
	return c.consumer.Start(ctx)
}

func (c *SagaKafkaConsumer) Close() error {
	c.logger.Info("closing saga kafka consumer")
	return c.consumer.Close()
}

func ParseUserDeletedEvent(data []byte) (*UserDeletedEvent, error) {
	var event UserDeletedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("unmarshal user.deleted: %w", err)
	}
	return &event, nil
}
