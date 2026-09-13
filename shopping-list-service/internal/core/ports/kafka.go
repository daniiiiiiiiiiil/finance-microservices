package ports

import "context"

type EventPublisher interface {
	Publish(ctx context.Context, eventType string, data interface{}) error
	SendShoppingDeletedEvent(ctx context.Context, userID int, deletedCount int, deletedIDs []int, imageKeys []string) error
	SendDeleteCompletedEvent(ctx context.Context, userID int, deletedShoppingCount int, deletedImagesCount int) error
	SendDeleteFailedEvent(ctx context.Context, userID int, reason string, err error, failedStep string) error
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
