package postgres

import (
	service_admin "backend/internal/features/admin/service"
	"fmt"

	"golang.org/x/net/context"
)

func (r *FinanceRepository) DeleteUserTransactions(ctx context.Context, userID int) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `DELETE FROM finance.transactions WHERE user_id = $1`
	cmdTag, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return 0, fmt.Errorf("delete transaction: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return 0, nil
	}

	return int(cmdTag.RowsAffected()), nil
}

func (r *FinanceRepository) GetMetrics(ctx context.Context) (service_admin.Metrics, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var metrics service_admin.Metrics

	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM finance.transactions`).Scan(&metrics.TotalTransactions)
	if err != nil {
		return service_admin.Metrics{}, fmt.Errorf("count transactions: %w", err)
	}

	err = r.pool.QueryRow(ctx, `
        SELECT COALESCE(SUM(
            CASE WHEN type_transaction = 'income' THEN amount ELSE -amount END
        ), 0)
        FROM finance.transactions
    `).Scan(&metrics.TotalBalance)
	if err != nil {
		return service_admin.Metrics{}, fmt.Errorf("total balance: %w", err)
	}

	return metrics, nil
}
