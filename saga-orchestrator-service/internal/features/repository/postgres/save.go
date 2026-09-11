package postgres

import (
	"context"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/ports"
)

func (r *SagaRepository) Save(ctx context.Context, saga *domain.Saga) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	INSERT INTO saga.sagas (
			saga_id, saga_type, user_id, status, current_step,
			total_steps, error, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
`
	model := sagaToModel(saga)

	var id int
	err := r.pool.QueryRow(ctx, query,
		model.SagaID,
		model.SagaType,
		model.UserID,
		model.Status,
		model.CurrentStep,
		model.TotalSteps,
		model.Error,
		model.Metadata,
		model.CreatedAt,
		model.UpdatedAt,
	).Scan(&id)
	if err != nil {
		return fmt.Errorf("cannot save saga to postgres: %w", err)
	}
	saga.ID = id
	return nil
}

func (r *SagaRepository) SaveTx(ctx context.Context, tx ports.Tx, saga *domain.Saga) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		INSERT INTO saga.sagas (
			saga_id, saga_type, user_id, status, current_step,
			total_steps, error, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
`
	model := sagaToModel(saga)

	var id int
	err := r.pool.QueryRow(ctx, query,
		model.SagaID,
		model.SagaType,
		model.UserID,
		model.Status,
		model.CurrentStep,
		model.TotalSteps,
		model.Error,
		model.Metadata,
		model.CreatedAt,
		model.UpdatedAt).Scan(&id)

	if err != nil {
		return fmt.Errorf("cannot save saga to postgres: %w", err)
	}
	saga.ID = id
	return nil
}
