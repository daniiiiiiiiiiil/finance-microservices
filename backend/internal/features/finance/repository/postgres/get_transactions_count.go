package postgres

import (
	"context"
	"fmt"
	"time"
)

func (r *FinanceRepository) GetTransactionsCount(
	ctx context.Context,
	userID int,
	transactionType, category *string,
	from, to *time.Time,
) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
        SELECT COUNT(*)
        FROM finance.transactions
        WHERE user_id = $1
    `

	args := []any{userID}
	argIndex := 2
	conditions := []string{}

	if transactionType != nil && *transactionType != "" {
		conditions = append(conditions, fmt.Sprintf("type_transaction = $%d", argIndex))
		args = append(args, *transactionType)
		argIndex++
	}

	if category != nil && *category != "" {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIndex))
		args = append(args, *category)
		argIndex++
	}

	if from != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, *from)
		argIndex++
	}

	if to != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, *to)
		argIndex++
	}

	if len(conditions) > 0 {
		query += " AND " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			query += " AND " + conditions[i]
		}
	}

	var total int
	err := r.pool.QueryRow(ctx, query, args...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("count transactions: %w", err)
	}

	return total, nil
}