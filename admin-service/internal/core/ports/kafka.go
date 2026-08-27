package ports

import (
	"context"

	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/internal/core/kafka"
)

type KafkaProducerInterface interface {
	SendEvent(ctx context.Context, eventType string, data interface{}) error
	Close() error
}

type KafkaConsumerInterface interface {
	Start(ctx context.Context) error
	RegisterHandler(eventType string, handler func(ctx context.Context, event kafka.Event) error)
	Close() error
}

type EventPublisherInterface interface {
	Publish(ctx context.Context, eventType string, data interface{}) error
}
