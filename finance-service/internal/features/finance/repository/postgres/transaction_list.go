package postgres

import (
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/domain"
	"golang.org/x/net/context"
)

func (r *FinanceRepository) GetTransactions(
	ctx context.Context,
	userID int,
	transactionType, category *string,
	from, to *time.Time,
	limit, offset int,
) ([]domain.Finance, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT id, version, type_transaction, amount, category, created_at, user_id
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

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get transactions: %w", err)
	}
	defer rows.Close()

	var models []FinanceModel
	for rows.Next() {
		var model FinanceModel
		err := rows.Scan(
			&model.ID,
			&model.Version,
			&model.TypeTransaction,
			&model.Amount,
			&model.Category,
			&model.CreatedAt,
			&model.UserID,
		)
		if err != nil {
			return nil, fmt.Errorf("scan transaction: %w", err)
		}
		models = append(models, model)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	transactions := make([]domain.Finance, len(models))
	for i, model := range models {
		transactions[i] = financeDomainFromModel(model)
	}

	return transactions, nil
}

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
