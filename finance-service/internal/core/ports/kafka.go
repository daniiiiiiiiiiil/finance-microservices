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

type EventHandler func(ctx context.Context, event Event) error

type Event struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Timestamp interface{} `json:"timestamp"`
	Data      []byte      `json:"data"`
}

type EventConsumer interface {
	RegisterHandler(eventType string, handler EventHandler) error
	Start(ctx context.Context) error
	Close() error
}
