package kafka

import (
	"context"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/kafka"
)

type UserEventPublisher struct {
	producer *kafka.Producer
}

func NewUserEventPublisher(producer *kafka.Producer) *UserEventPublisher {
	return &UserEventPublisher{producer: producer}
}

func (p *UserEventPublisher) Publish(ctx context.Context, eventType string, data interface{}) error {
	if err := p.producer.SendEvent(ctx, eventType, data); err != nil {
		return fmt.Errorf("publish event: %w", err)
	}
	return nil
}
