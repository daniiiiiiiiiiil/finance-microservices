package ports

import (
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/repository/postgres/pool"
	"golang.org/x/net/context"
)

type OutboxRepositoryInterface interface {
	SaveTx(ctx context.Context, tx pool.Tx, event domain.OutboxEvent) error
	Save(ctx context.Context, event domain.OutboxEvent) error
	GetPending(ctx context.Context, limit int) ([]domain.OutboxEvent, error)
	MarkProcessed(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id string, errMsg string) error
}
