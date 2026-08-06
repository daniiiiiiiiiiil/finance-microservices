package postgres

import (
	"backend/internal/features/admin/service"
	"context"
	"fmt"
)

type Metrics struct {
	TotalUsers        int     `json:"total_users"`
	TotalTransactions int     `json:"total_transactions"`
	TotalBalance      float64 `json:"total_balance"`
}

func (r *AdminRepository) GetMetrics(ctx context.Context) (service.Metrics, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var metrics service.Metrics

	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM finance.users`).Scan(&metrics.TotalUsers)
	if err != nil {
		return service.Metrics{}, fmt.Errorf("count users: %w", err)
	}

	err = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM finance.transactions`).Scan(&metrics.TotalTransactions)
	if err != nil {
		return service.Metrics{}, fmt.Errorf("count transactions: %w", err)
	}

	err = r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(
			CASE WHEN type_transaction = 'income' THEN amount ELSE -amount END
		), 0)
		FROM finance.transactions
	`).Scan(&metrics.TotalBalance)
	if err != nil {
		return service.Metrics{}, fmt.Errorf("total balance: %w", err)
	}

	return metrics, nil
}
