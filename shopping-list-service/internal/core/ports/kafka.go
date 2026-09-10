package ports

import "context"

type EventConsumer interface {
	Start(ctx context.Context) error
	Close() error
}

type EventPublisher interface {
	Publish(ctx context.Context, eventType string, data interface{}) error

	SendShoppingDeletedEvent(ctx context.Context, userID int, deletedCount int, deletedIDs []int, imageKeys []string) error
	SendDeleteCompletedEvent(ctx context.Context, userID int, deletedShoppingCount int, deletedImagesCount int) error
	SendDeleteFailedEvent(ctx context.Context, userID int, reason string, err error, failedStep string) error
}

type SagaOrchestrator interface {
	HandleUserDeleted(ctx context.Context, userID int) error
}
