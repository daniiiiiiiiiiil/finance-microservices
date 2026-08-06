package postgres

import (
	errors_core "backend/internal/core/errors"
	"context"
	"fmt"
)

func (r *FinanceRepository) DeleteTransaction(ctx context.Context, id int) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `DELETE FROM finance.transactions WHERE id = $1`

	cmdTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete transaction: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("transaction with id %d: %w", id, errors_core.ErrNotFound)
	}

	return nil
}
