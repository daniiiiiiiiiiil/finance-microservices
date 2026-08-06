package postgres

import (
	"backend/internal/core/domain"
	errors_core "backend/internal/core/errors"
	"backend/internal/core/repository/postgres/pool"
	"context"
	"errors"
	"fmt"
)

func (r *FinanceRepository) GetTransaction(ctx context.Context, id int) (domain.Finance, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT id, version, type_transaction, amount, category, created_at, user_id
		FROM finance.transactions
		WHERE id = $1
	`

	var model FinanceModel
	err := r.pool.QueryRow(ctx, query, id).Scan(
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
			return domain.Finance{}, fmt.Errorf("transaction with id %d: %w", id, errors_core.ErrNotFound)
		}
		return domain.Finance{}, fmt.Errorf("get transaction: %w", err)
	}

	return financeDomainFromModel(model), nil
}
