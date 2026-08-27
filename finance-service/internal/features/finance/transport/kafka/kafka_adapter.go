package kafka

import (
	"context"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/kafka"
)

type FinanceEventPublisher struct {
	producer *kafka.Producer
}

func NewFinanceEventPublisher(producer *kafka.Producer) *FinanceEventPublisher {
	return &FinanceEventPublisher{producer: producer}
}

func (p *FinanceEventPublisher) Publish(ctx context.Context, eventType string, data interface{}) error {
	if err := p.producer.SendEvent(ctx, eventType, data); err != nil {
		return fmt.Errorf("publish event: %w", err)
	}
	return nil
}
