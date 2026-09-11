package ports

import (
	"context"

	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/domain"
)

type SagaOrchestrator interface {
	StartSaga(ctx context.Context, saga *domain.Saga) error
	HandleUserDeleted(ctx context.Context, userID int) error
	GetSaga(ctx context.Context, sagaID string) (*domain.Saga, error)
	GetActiveSagas(ctx context.Context) ([]*domain.Saga, error)
}

type SagaManager interface {
	StartSaga(ctx context.Context, saga *domain.Saga) error
	ExecuteStep(ctx context.Context, saga *domain.Saga, step *domain.Step) error
	Compensate(ctx context.Context, saga *domain.Saga) error
	Shutdown(ctx context.Context) error
}

type SagaDefinition struct {
	Type  domain.SagaType
	Steps []SagaStepDefinition
	Build func(ctx context.Context, saga *domain.Saga) error
}

type SagaStepDefinition struct {
	Name       string
	Order      int
	Action     func(ctx context.Context, saga *domain.Saga) error
	Compensate func(ctx context.Context, saga *domain.Saga) error
}

type SagaRegistry interface {
	Register(sagaType domain.SagaType, definition SagaDefinition) error
	Get(sagaType domain.SagaType) (SagaDefinition, error)
	GetAll() []SagaDefinition
}
