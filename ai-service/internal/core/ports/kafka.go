package ports

import "context"

type EventPublisher interface {
	Publish(ctx context.Context, eventType string, data any) error
}
