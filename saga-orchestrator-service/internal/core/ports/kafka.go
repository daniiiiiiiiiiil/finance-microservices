package ports

import "context"

type EventPublisher interface {
	Publish(ctx context.Context, eventType string, data interface{}) error
	SendUserMarkDeleting(ctx context.Context, userID int) error
	SendDeleteShoppingData(ctx context.Context, userID int) error
	SendDeleteFinanceTransactions(ctx context.Context, userID int) error
	SendUserFinalizeDelete(ctx context.Context, userID int) error
	SendUserDeleteCompleted(ctx context.Context, userID int, sagaID string) error
	SendUserDeleteFailed(ctx context.Context, userID int, sagaID string, reason string, errMsg string) error
}

type EventConsumer interface {
	RegisterHandler(eventType string, handler EventHandler) error
	Start(ctx context.Context) error
	Close() error
}

type EventHandler func(ctx context.Context, event Event) error
type Event struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Timestamp interface{} `json:"timestamp"`
	Data      []byte      `json:"data"`
}
