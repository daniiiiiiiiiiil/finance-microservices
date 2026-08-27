package kafka

import (
	"context"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/internal/core/kafka"
)

type AdminEventPublisher struct {
	producer *kafka.Producer
}

func NewAdminEventPublisher(producer *kafka.Producer) *AdminEventPublisher {
	return &AdminEventPublisher{producer: producer}
}

func (p *AdminEventPublisher) Publish(ctx context.Context, eventType string, data interface{}) error {
	if err := p.producer.SendEvent(ctx, eventType, data); err != nil {
		return fmt.Errorf("publish event: %w", err)
	}
	return nil
}
