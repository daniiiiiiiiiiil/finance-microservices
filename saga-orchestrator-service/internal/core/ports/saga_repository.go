package ports

import (
	"context"

	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/repository/postgres/pool"
)

type SagaRepository interface {
	Save(ctx context.Context, saga *domain.Saga) error
	SaveTx(ctx context.Context, tx pool.Tx, saga *domain.Saga) error
	Update(ctx context.Context, saga *domain.Saga) error
	UpdateStatus(ctx context.Context, sagaID int, status domain.Status, errMsg string) error
	UpdateStep(ctx context.Context, step *domain.Step) error
	SaveCompensation(ctx context.Context, compensation *domain.Compensation) error
	UpdateCompensation(ctx context.Context, compensation *domain.Compensation) error
	GetByID(ctx context.Context, id int) (*domain.Saga, error)
	GetBySagaID(ctx context.Context, sagaID string) (*domain.Saga, error)
	GetByUserID(ctx context.Context, userID int) ([]*domain.Saga, error)
	GetActive(ctx context.Context) ([]*domain.Saga, error)
	GetByStatus(ctx context.Context, status domain.Status) ([]*domain.Saga, error)
	Delete(ctx context.Context, id int) error
}
