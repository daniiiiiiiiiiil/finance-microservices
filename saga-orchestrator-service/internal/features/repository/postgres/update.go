package postgres

import (
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/domain"
	"golang.org/x/net/context"
)

func (r *SagaRepository) Update(ctx context.Context, saga *domain.Saga) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query :=
		`
		UPDATE saga.sagas SET
		    	status = $1,
		    	current_step = $2,
		    	error = $3,
		    	metadata = $4,
		    	updated_at = $5,
		    	completed_at = $6,
		WHERE saga.id = $7
`
	model := sagaToModel(saga)

	_, err := r.pool.Exec(ctx, query,
		model.Status,
		model.CurrentStep,
		model.Error,
		model.Metadata,
		time.Now(),
		model.CompletedAt,
		model.ID)
	if err != nil {
		return fmt.Errorf("error while updating saga: %w", err)
	}
	return nil
}

func (r *SagaRepository) UpdateStatus(ctx context.Context, sagaID int, status domain.Status, errMsg string) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var errPtr *string
	if errMsg != "" {
		errPtr = &errMsg
	}

	query := `
		UPDATE saga.sagas SET
			status = $1,
			error = $2,
			updated_at = $3,
			completed_at = CASE 
				WHEN $1 IN ('completed', 'failed', 'compensated') THEN $3 
				ELSE completed_at 
			END
		WHERE id = $4
	`

	_, err := r.pool.Exec(ctx, query,
		status.String(),
		errPtr,
		time.Now(),
		sagaID,
	)

	if err != nil {
		return fmt.Errorf("update saga status: %w", err)
	}
	return nil
}

func (r *SagaRepository) UpdateStep(ctx context.Context, step *domain.Step) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		UPDATE saga.saga_steps SET
			status = $1,
			error = $2,
			metadata = $3,
			started_at = $4,
			completed_at = $5,
			updated_at = $6
		WHERE id = $7
`
	model := stepToModel(step)

	_, err := r.pool.Exec(ctx, query,
		model.Status,
		model.Error,
		model.Metadata,
		model.StartedAt,
		model.CompletedAt,
		time.Now(),
		model.ID,
	)

	if err != nil {
		return fmt.Errorf("update step: %w", err)
	}
	return nil
}

func (r *SagaRepository) UpdateCompensation(ctx context.Context, comp *domain.Compensation) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		UPDATE saga.saga_compensations SET
			status = $1,
			error = $2,
			started_at = $3,
			completed_at = $4,
			updated_at = $5
		WHERE id = $6
	`

	model := compensationToModel(comp)

	_, err := r.pool.Exec(ctx, query,
		model.Status,
		model.Error,
		model.StartedAt,
		model.CompletedAt,
		time.Now(),
		model.ID,
	)

	if err != nil {
		return fmt.Errorf("update compensation: %w", err)
	}
	return nil
}
