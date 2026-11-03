package kafka

import (
	"context"

	corekafka "github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/core/kafka"
	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/core/ports"
)

type EventPublisher struct {
	producer *corekafka.Producer
}

func NewEventPublisher(p *corekafka.Producer) *EventPublisher {
	return &EventPublisher{producer: p}
}

var _ ports.EventPublisher = (*EventPublisher)(nil)

func (p *EventPublisher) Publish(ctx context.Context, eventType string, data any) error {
	return p.producer.SendEvent(ctx, eventType, data)
}
