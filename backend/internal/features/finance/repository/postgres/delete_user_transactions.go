package postgres

import (
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
