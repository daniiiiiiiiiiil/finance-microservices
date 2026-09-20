package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/repository/postgres/pool"
)

func (r *SagaRepository) SaveOutboxTx(ctx context.Context, tx pool.Tx, event domain.OutboxEvent) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		INSERT INTO saga.outbox_messages (
			aggregate_id, aggregate_type, event_type, payload, status, max_retries
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var id int64
	err := tx.QueryRow(ctx, query,
		event.AggregateID,
		event.AggregateType,
		event.EventType,
		event.EventPayload,
		"pending",
		3).Scan(&id)
	if err != nil {
		return fmt.Errorf("cannot save outbox tx: %w", err)
	}
	return nil
}

func (r *SagaRepository) SaveOutbox(ctx context.Context, event domain.OutboxEvent) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		INSERT INTO saga.outbox_messages (
			aggregate_id, aggregate_type, event_type, payload, status, max_retries
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var id int64
	err := r.pool.QueryRow(ctx, query,
		event.AggregateID,
		event.AggregateType,
		event.EventType,
		event.EventPayload,
		"pending",
		3).Scan(&id)
	if err != nil {
		return fmt.Errorf("cannot save outbox tx: %w", err)
	}
	return nil
}

func (r *SagaRepository) GetPendingOutbox(ctx context.Context, limit int) ([]domain.OutboxEvent, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT id, aggregate_id, aggregate_type, event_type, payload,
		       created_at, processed_at, status, last_error, retry_count
		FROM saga.outbox_messages
		WHERE status = 'pending' AND  retry_count < max_retries
		ORDER BY id ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("cannot get pending outbox tx: %w", err)
	}
	defer rows.Close()

	var events []domain.OutboxEvent
	for rows.Next() {
		var e domain.OutboxEvent
		var id int64
		var processedAt *time.Time
		var lastError *string
		var retryCount int

		err := rows.Scan(
			&id,
			&e.AggregateID,
			&e.AggregateType,
			&e.EventType,
			&e.EventPayload,
			&e.CreatedAt,
			&processedAt,
			&e.Status,
			&lastError,
			&retryCount)
		if err != nil {
			return nil, fmt.Errorf("cannot get pending outbox tx: %w", err)
		}

		e.ID = fmt.Sprintf("%d", id)
		if processedAt != nil {
			e.ProcessAt = *processedAt
		}
		if lastError != nil {
			e.ErrorMessage = *lastError
		}
		e.RetryCount = retryCount
		events = append(events, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot get pending outbox rows: %w", err)
	}
	return events, nil
}

func (r *SagaRepository) MarkOutboxProcessed(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		UPDATE saga.outbox_messages
		SET status = 'processed', processed_at = NOW(),version = version + 1
		WHERE id = $1 AND status = 'pending'
`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("cannot mark outbox processed tx: %w", err)
	}
	return nil
}

func (r *SagaRepository) MarkOutboxFailed(ctx context.Context, id string, errMsg string) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		UPDATE saga.outbox_messages
		SET status = 'failed',last_error = $2,retry_count = retry_count + 1,version = version + 1
		WHERE id = $1
`
	_, err := r.pool.Exec(ctx, query, id, errMsg)
	if err != nil {
		return fmt.Errorf("cannot mark outbox failed tx: %w", err)
	}
	return nil
}
