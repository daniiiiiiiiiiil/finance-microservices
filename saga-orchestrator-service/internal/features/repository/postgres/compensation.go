package postgres

import (
	"context"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/domain"
)

func (r *SagaRepository) SaveCompensation(ctx context.Context, comp *domain.Compensation) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		INSERT INTO saga.saga_compensations(
		        saga_id, step_name, compensation_data, status,
			error, started_at, completed_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
`
	model := compensationToModel(comp)

	var id int
	err := r.pool.QueryRow(ctx, query,
		model.SagaID,
		model.StepName,
		model.CompensationData,
		model.Status,
		model.Error,
		model.StartedAt,
		model.CompletedAt,
		model.CreatedAt,
		model.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("save compensation: %w", err)
	}

	comp.ID = id
	return nil
}
