package ports

import "golang.org/x/net/context"

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
