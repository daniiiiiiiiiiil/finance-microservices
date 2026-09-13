package ports

import "context"

type KafkaProducerInterface interface {
	SendEvent(ctx context.Context, eventType string, data interface{}) error
	Close() error
}

type EventPublisherInterface interface {
	Publish(ctx context.Context, eventType string, data interface{}) error
}
type UserEvent struct {
	UserID   int    `json:"user_id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	IsAdmin  bool   `json:"is_admin"`
	Status   string `json:"status"`
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
