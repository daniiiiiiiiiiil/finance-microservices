package postgres

import (
	"backend/internal/core/domain"
	"backend/internal/core/repository/postgres/pool"
	"context"
	"fmt"
)

func (r *FinanceRepository) CreateTransactionTx(ctx context.Context, tx pool.Tx, transaction domain.Finance) (domain.Finance, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		INSERT INTO finance.transactions (type_transaction, amount, category, created_at, user_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, version, type_transaction, amount, category, created_at, user_id
	`

	var model FinanceModel
	err := tx.QueryRow(ctx, query,
		transaction.TypeTransaction,
		transaction.Amount,
		transaction.Category,
		transaction.CreatedAt,
		transaction.UserID,
	).Scan(
		&model.ID,
		&model.Version,
		&model.TypeTransaction,
		&model.Amount,
		&model.Category,
		&model.CreatedAt,
		&model.UserID,
	)
	if err != nil {
		return domain.Finance{}, fmt.Errorf("insert transaction: %w", err)
	}

	return financeDomainFromModel(model), nil
}
