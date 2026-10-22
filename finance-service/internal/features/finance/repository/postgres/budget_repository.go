package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/repository/postgres/pool"
	errors_my "github.com/daniiiiiiiiiiil/finance-microservices/finance-service/pkg/errors"
)

func (r *FinanceRepository) CreateBudgetTx(
	ctx context.Context,
	tx pool.Tx,
	b domain.Budget,
) (domain.Budget, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		INSERT INTO finance.budgets
			(user_id, category, amount_limit, period, period_start, period_end, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, version, user_id, category, amount_limit, period,
		          period_start, period_end, created_at, updated_at
	`
	var m BudgetModel
	err := tx.QueryRow(ctx, query,
		b.UserID, b.Category, b.AmountLimit, string(b.Period), b.PeriodStart, b.PeriodEnd,
	).Scan(
		&m.ID, &m.Version, &m.UserID, &m.Category, &m.AmountLimit, &m.Period,
		&m.PeriodStart, &m.PeriodEnd, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return domain.Budget{}, fmt.Errorf("insert budget: %w", err)
	}
	return budgetDomainFromModel(m), nil
}

func (r *FinanceRepository) GetBudget(ctx context.Context, id int) (domain.Budget, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT id, version, user_id, category, amount_limit, period,
		       period_start, period_end, created_at, updated_at
		FROM finance.budgets
		WHERE id = $1
	`
	var m BudgetModel
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&m.ID, &m.Version, &m.UserID, &m.Category, &m.AmountLimit, &m.Period,
		&m.PeriodStart, &m.PeriodEnd, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pool.ErrNoRows) {
			return domain.Budget{}, fmt.Errorf("budget with id %d: %w", id, errors_my.ErrNotFound)
		}
		return domain.Budget{}, fmt.Errorf("get budget: %w", err)
	}
	return budgetDomainFromModel(m), nil
}

func (r *FinanceRepository) ListBudgets(ctx context.Context, userID int) ([]domain.Budget, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT id, version, user_id, category, amount_limit, period,
		       period_start, period_end, created_at, updated_at
		FROM finance.budgets
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list budgets: %w", err)
	}
	defer rows.Close()

	var budgets []domain.Budget
	for rows.Next() {
		var m BudgetModel
		if err := rows.Scan(
			&m.ID, &m.Version, &m.UserID, &m.Category, &m.AmountLimit, &m.Period,
			&m.PeriodStart, &m.PeriodEnd, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan budget: %w", err)
		}
		budgets = append(budgets, budgetDomainFromModel(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return budgets, nil
}

func (r *FinanceRepository) UpdateBudget(ctx context.Context, b domain.Budget) (domain.Budget, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		UPDATE finance.budgets SET
			category     = $1,
			amount_limit = $2,
			period       = $3,
			period_start = $4,
			period_end   = $5,
			version      = version + 1,
			updated_at   = NOW()
		WHERE id = $6 AND version = $7
		RETURNING id, version, user_id, category, amount_limit, period,
		          period_start, period_end, created_at, updated_at
	`
	var m BudgetModel
	err := r.pool.QueryRow(ctx, query,
		b.Category, b.AmountLimit, string(b.Period), b.PeriodStart, b.PeriodEnd,
		b.ID, b.Version,
	).Scan(
		&m.ID, &m.Version, &m.UserID, &m.Category, &m.AmountLimit, &m.Period,
		&m.PeriodStart, &m.PeriodEnd, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pool.ErrNoRows) {
			return domain.Budget{}, fmt.Errorf("budget with id='%d' concurrently accessed: %w", b.ID, errors_my.ErrConflict)
		}
		return domain.Budget{}, fmt.Errorf("update budget: %w", err)
	}
	return budgetDomainFromModel(m), nil
}

func (r *FinanceRepository) DeleteBudget(ctx context.Context, id int) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `DELETE FROM finance.budgets WHERE id = $1`
	cmdTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete budget: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("budget with id %d: %w", id, errors_my.ErrNotFound)
	}
	return nil
}

func (r *FinanceRepository) GetBudgetSpent(ctx context.Context, b domain.Budget) (float64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM finance.transactions
		WHERE user_id = $1
		  AND category = $2
		  AND type_transaction = 'expense'
		  AND created_at >= $3
		  AND created_at <  $4
	`
	var spent float64
	if err := r.pool.QueryRow(ctx, query, b.UserID, b.Category, b.PeriodStart, b.PeriodEnd).Scan(&spent); err != nil {
		return 0, fmt.Errorf("get budget spent: %w", err)
	}
	return spent, nil
}

func (r *FinanceRepository) FindActiveBudget(
	ctx context.Context,
	userID int,
	category string,
	period domain.BudgetPeriod,
	now time.Time,
) (domain.Budget, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT id, version, user_id, category, amount_limit, period,
		       period_start, period_end, created_at, updated_at
		FROM finance.budgets
		WHERE user_id = $1
		  AND category = $2
		  AND period = $3
		  AND period_start <= $4
		  AND period_end   >  $4
		ORDER BY created_at DESC
		LIMIT 1
	`
	var m BudgetModel
	err := r.pool.QueryRow(ctx, query, userID, category, string(period), now).Scan(
		&m.ID, &m.Version, &m.UserID, &m.Category, &m.AmountLimit, &m.Period,
		&m.PeriodStart, &m.PeriodEnd, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pool.ErrNoRows) {
			return domain.Budget{}, false, nil
		}
		return domain.Budget{}, false, fmt.Errorf("find active budget: %w", err)
	}
	return budgetDomainFromModel(m), true, nil
}
