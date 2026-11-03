package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/core/repository/postgres/pool"
)

func (r *PlanRepository) Save(ctx context.Context, plan *domain.SavedPlan) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	model, err := planFromDomain(plan)
	if err != nil {
		return fmt.Errorf("marshal plan: %w", err)
	}

	query := `
		INSERT INTO ai.shopping_plans
			(user_id, request_id, text, address, plan, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, created_at
	`
	err = r.pool.QueryRow(ctx, query,
		model.UserID,
		model.RequestID,
		model.Text,
		model.Address,
		model.PlanJSON,
	).Scan(&plan.ID, &plan.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert plan: %w", err)
	}
	return nil
}

func (r *PlanRepository) GetByID(ctx context.Context, id int) (*domain.SavedPlan, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT id, user_id, request_id, text, address, plan, created_at
		FROM ai.shopping_plans
		WHERE id = $1
	`
	var m PlanModel
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&m.ID, &m.UserID, &m.RequestID, &m.Text, &m.Address, &m.PlanJSON, &m.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pool.ErrNoRows) {
			return nil, fmt.Errorf("plan with id %d: %w", id, domain.ErrPlanNotFound)
		}
		return nil, fmt.Errorf("get plan by id: %w", err)
	}
	return m.ToDomain()
}

func (r *PlanRepository) GetByRequestID(ctx context.Context, requestID string) (*domain.SavedPlan, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT id, user_id, request_id, text, address, plan, created_at
		FROM ai.shopping_plans
		WHERE request_id = $1
	`
	var m PlanModel
	err := r.pool.QueryRow(ctx, query, requestID).Scan(
		&m.ID, &m.UserID, &m.RequestID, &m.Text, &m.Address, &m.PlanJSON, &m.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pool.ErrNoRows) {
			return nil, fmt.Errorf("plan with request_id %s: %w", requestID, domain.ErrPlanNotFound)
		}
		return nil, fmt.Errorf("get plan by request_id: %w", err)
	}
	return m.ToDomain()
}

func (r *PlanRepository) ListByUser(ctx context.Context, userID int, limit, offset int) ([]domain.PlanSummary, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT
			id, request_id, text, address,
			COALESCE(jsonb_array_length(plan->'items'), 0) AS items_count,
			COALESCE(jsonb_array_length(plan->'stores'), 0) AS stores_count,
			COALESCE((plan->'cost'->>'total')::numeric, 0) AS total_cost,
			COALESCE(plan->'recommendation'->>'mode', '') AS recommended,
			created_at
		FROM ai.shopping_plans
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list plans: %w", err)
	}
	defer rows.Close()

	var out []domain.PlanSummary
	for rows.Next() {
		var s domain.PlanSummary
		if err := rows.Scan(
			&s.ID, &s.RequestID, &s.Text, &s.Address,
			&s.ItemsCount, &s.StoresCount, &s.TotalCost, &s.Recommended,
			&s.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan plan summary: %w", err)
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return out, nil
}

func (r *PlanRepository) CountByUser(ctx context.Context, userID int) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM ai.shopping_plans WHERE user_id = $1`, userID,
	).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("count plans: %w", err)
	}
	return total, nil
}

func (r *PlanRepository) Delete(ctx context.Context, id int) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	tag, err := r.pool.Exec(ctx, `DELETE FROM ai.shopping_plans WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete plan: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("plan with id %d: %w", id, domain.ErrPlanNotFound)
	}
	return nil
}
func (r *PlanRepository) DeleteByUser(ctx context.Context, userID int) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	_, err := r.pool.Exec(ctx, `DELETE FROM ai.shopping_plans WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("delete plans by user %d: %w", userID, err)
	}
	return nil
}
