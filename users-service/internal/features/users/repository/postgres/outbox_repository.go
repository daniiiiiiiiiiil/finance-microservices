package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/repository/postgres/pool"
)

type OutboxRepository struct {
	pool pool.Pool
}

func NewOutboxRepository(pool pool.Pool) *OutboxRepository {
	return &OutboxRepository{pool: pool}
}

func (r *OutboxRepository) SaveTx(ctx context.Context, tx pool.Tx, event domain.OutboxEvent) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		INSERT INTO outbox.outbox_messages (
			aggregate_id, aggregate_type, event_type, payload, status, max_retries
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	var id int64
	err := tx.QueryRow(ctx, query,
		event.AggregateID,
		event.AggregateType,
		event.EventType,
		event.EventPayload,
		"pending",
		3,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("save outbox message in tx: %w", err)
	}
	return nil
}

func (r *OutboxRepository) Save(ctx context.Context, event domain.OutboxEvent) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		INSERT INTO outbox.outbox_messages (
			aggregate_id, aggregate_type, event_type, payload, status, max_retries
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	var id int64
	err := r.pool.QueryRow(ctx, query,
		event.AggregateID,
		event.AggregateType,
		event.EventType,
		event.EventPayload,
		"pending",
		3,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("save outbox message: %w", err)
	}
	return nil
}

func (r *OutboxRepository) GetPending(ctx context.Context, limit int) ([]domain.OutboxEvent, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT id, aggregate_id, aggregate_type, event_type, payload, 
		       created_at, processed_at, status, last_error, retry_count
		FROM outbox.outbox_messages
		WHERE status = 'pending' AND retry_count < max_retries
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("get pending messages: %w", err)
	}
	defer rows.Close()

	var events []domain.OutboxEvent
	for rows.Next() {
		var event domain.OutboxEvent
		var id int64
		var processedAt *time.Time
		var lastError *string
		var retryCount int

		err := rows.Scan(
			&id,
			&event.AggregateID,
			&event.AggregateType,
			&event.EventType,
			&event.EventPayload,
			&event.CreatedAt,
			&processedAt,
			&event.Status,
			&lastError,
			&retryCount,
		)
		if err != nil {
			return nil, fmt.Errorf("scan outbox message: %w", err)
		}

		event.ID = fmt.Sprintf("%d", id)
		if processedAt != nil {
			event.ProcessedAt = *processedAt
		}
		if lastError != nil {
			event.ErrorMessage = *lastError
		}
		event.RetryCount = retryCount

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return events, nil
}

func (r *OutboxRepository) MarkProcessed(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		UPDATE outbox.outbox_messages 
		SET status = 'processed', processed_at = NOW(), version = version + 1
		WHERE id = $1 AND status = 'pending'`

	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("mark message processed: %w", err)
	}
	return nil
}

func (r *OutboxRepository) MarkFailed(ctx context.Context, id string, errMsg string) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		UPDATE outbox.outbox_messages 
		SET status = 'failed', last_error = $2, retry_count = retry_count + 1, version = version + 1
		WHERE id = $1`

	_, err := r.pool.Exec(ctx, query, id, errMsg)
	if err != nil {
		return fmt.Errorf("mark message failed: %w", err)
	}
	return nil
}
