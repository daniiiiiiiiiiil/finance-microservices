package ports

import (
	"context"

	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/domain"
)

type OutboxRepository interface {
	SaveOutboxTx(ctx context.Context, tx Tx, event domain.OutboxEvent) error
	SaveOutbox(ctx context.Context, event domain.OutboxEvent) error
	GetPendingOutbox(ctx context.Context, limit int) ([]domain.OutboxEvent, error)
	MarkOutboxProcessed(ctx context.Context, id string) error
	MarkOutboxFailed(ctx context.Context, id string, errMsg string) error
}
