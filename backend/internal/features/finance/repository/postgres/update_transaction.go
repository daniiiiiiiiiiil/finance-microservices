package postgres

import (
	"backend/internal/core/domain"
	errors_core "backend/internal/core/errors"
	"backend/internal/core/repository/postgres/pool"
	"context"
	"errors"
	"fmt"
)

func (r *FinanceRepository) UpdateTransaction(ctx context.Context, transaction domain.Finance) (domain.Finance, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		UPDATE finance.transactions SET
			type_transaction = $1,
			amount = $2,
			category = $3,
			version = version + 1
		WHERE id = $4 AND version = $5
		RETURNING id, version, type_transaction, amount, category, created_at, user_id
	`

	var model FinanceModel
	err := r.pool.QueryRow(ctx, query,
		transaction.TypeTransaction,
		transaction.Amount,
		transaction.Category,
		transaction.ID,
		transaction.Version,
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
		if errors.Is(err, pool.ErrNoRows) {
			return domain.Finance{}, fmt.Errorf("transaction with id='%d' concurrently accessed: %w", transaction.ID, errors_core.ErrConflict)
		}
		return domain.Finance{}, fmt.Errorf("update transaction: %w", err)
	}

	return financeDomainFromModel(model), nil
}
