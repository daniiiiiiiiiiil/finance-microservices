package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/domain"
)

func (r *SagaRepository) GetByID(ctx context.Context, id int) (*domain.Saga, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT id,saga_id,saga_type,user_id,status,current_step,
				total_steps,error,metadata,created_at,updated_at,completed_at
		FROM saga.sagas
		WHERE id = $1
`
	var model SagaModel
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&model.ID,
		&model.SagaID,
		&model.SagaType,
		&model.UserID,
		&model.Status,
		&model.CurrentStep,
		&model.TotalSteps,
		&model.Error,
		&model.Metadata,
		&model.CreatedAt,
		&model.UpdatedAt,
		&model.CompletedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("saga with id %d not found: %w", id, errors.New("not found"))
		}
		return nil, fmt.Errorf("get saga by id: %w", err)
	}

	saga := sagaFromModel(&model)

	steps, err := r.getStepsBySagaID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get saga by id %d: %w", id, err)
	}
	saga.Steps = steps
	return saga, nil
}

func (r *SagaRepository) GetBySagaID(ctx context.Context, sagaID string) (*domain.Saga, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT id,saga_id,saga_type,user_id,status,current_step,
				total_steps,error,metadata,created_at,updated_at,completed_at
		FROM saga.sagas
		WHERE id = $1
`

	var model SagaModel
	err := r.pool.QueryRow(ctx, query, sagaID).Scan(
		&model.ID,
		&model.SagaID,
		&model.SagaType,
		&model.UserID,
		&model.Status,
		&model.CurrentStep,
		&model.TotalSteps,
		&model.Error,
		&model.Metadata,
		&model.CreatedAt,
		&model.UpdatedAt,
		&model.CompletedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("saga with id %s not found: %w", sagaID, errors.New("not found"))
		}
		return nil, fmt.Errorf("get saga by id %s: %w", sagaID, err)
	}

	saga := sagaFromModel(&model)
	steps, err := r.getStepsBySagaID(ctx, model.ID)
	if err != nil {
		return nil, fmt.Errorf("get saga by id %d: %w", model.ID, err)
	}
	saga.Steps = steps
	return saga, nil
}

func (r *SagaRepository) GetByUserID(ctx context.Context, userID int) ([]*domain.Saga, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT id, saga_id, saga_type, user_id, status, current_step,
			total_steps, error, metadata, created_at, updated_at, completed_at
		FROM saga.sagas
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get saga by user id %d: %w", userID, err)
	}
	defer rows.Close()
	var sagas []*domain.Saga
	for rows.Next() {
		var model SagaModel
		err := rows.Scan(
			&model.ID,
			&model.SagaID,
			&model.SagaType,
			&model.UserID,
			&model.Status,
			&model.CurrentStep,
			&model.TotalSteps,
			&model.Error,
			&model.Metadata,
			&model.CreatedAt,
			&model.UpdatedAt,
			&model.CompletedAt)
		if err != nil {
			return nil, fmt.Errorf("get saga by user id %d: %w", userID, err)
		}
		sagas = append(sagas, sagaFromModel(&model))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get saga by user id %d: %w", userID, err)
	}

	return sagas, nil
}

func (r *SagaRepository) GetActive(ctx context.Context) ([]*domain.Saga, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT id, saga_id, saga_type, user_id, status, current_step,
			total_steps, error, metadata, created_at, updated_at, completed_at
		FROM saga.sagas
		WHERE status IN ('pending', 'in_progress', 'compensating')
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get saga active %d: %w", 1, err)
	}
	defer rows.Close()

	var sagas []*domain.Saga
	for rows.Next() {
		var model SagaModel
		err := rows.Scan(
			&model.ID,
			&model.SagaID,
			&model.SagaType,
			&model.UserID,
			&model.Status,
			&model.CurrentStep,
			&model.TotalSteps,
			&model.Error,
			&model.Metadata,
			&model.CreatedAt,
			&model.UpdatedAt,
			&model.CompletedAt)
		if err != nil {
			return nil, fmt.Errorf("get saga active %d: %w", model.ID, err)
		}
		sagas = append(sagas, sagaFromModel(&model))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get saga active: %w", err)
	}
	return sagas, nil
}

func (r *SagaRepository) GetByStatus(ctx context.Context, status domain.Status) ([]*domain.Saga, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
			SELECT id,saga_id,saga_type,user_id,status,current_step,
				total_steps,error,metadata,created_at,updated_at,completed_at
			FROM saga.sagas
			WHERE status = $1
			ORDER BY created_at ASC
`
	rows, err := r.pool.Query(ctx, query, status)
	if err != nil {
		return nil, fmt.Errorf("get saga by status: %w", err)
	}
	defer rows.Close()

	var sagas []*domain.Saga
	for rows.Next() {
		var model SagaModel
		err := rows.Scan(
			&model.ID,
			&model.SagaID,
			&model.SagaType,
			&model.UserID,
			&model.Status,
			&model.CurrentStep,
			&model.TotalSteps,
			&model.Error,
			&model.Metadata,
			&model.CreatedAt,
			&model.UpdatedAt,
			&model.CompletedAt)
		if err != nil {
			return nil, fmt.Errorf("get saga by status: %w", err)
		}
		sagas = append(sagas, sagaFromModel(&model))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get saga by status: %w", err)
	}
	return sagas, nil
}

func (r *SagaRepository) getStepsBySagaID(ctx context.Context, sagaID int) ([]*domain.Step, error) {
	query := `
		SELECT id, saga_id, step_name, step_order, status, error,
			metadata, started_at, completed_at, created_at, updated_at
		FROM saga.saga_steps
		WHERE saga_id = $1
		ORDER BY step_order ASC
	`

	rows, err := r.pool.Query(ctx, query, sagaID)
	if err != nil {
		return nil, fmt.Errorf("get steps: %w", err)
	}
	defer rows.Close()

	var steps []*domain.Step
	for rows.Next() {
		var model StepModel
		err := rows.Scan(
			&model.ID,
			&model.SagaID,
			&model.StepName,
			&model.StepOrder,
			&model.Status,
			&model.Error,
			&model.Metadata,
			&model.StartedAt,
			&model.CompletedAt,
			&model.CreatedAt,
			&model.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan step: %w", err)
		}
		steps = append(steps, stepFromModel(&model))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return steps, nil
}
