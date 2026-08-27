package ports

import (
	"context"
)

type EventPublisherInterface interface {
	Publish(ctx context.Context, eventType string, data interface{}) error
}

type KafkaProducerInterface interface {
	SendEvent(ctx context.Context, eventType string, data interface{}) error
	Close() error
}
