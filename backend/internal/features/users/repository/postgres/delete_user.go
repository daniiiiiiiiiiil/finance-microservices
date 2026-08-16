package postgres

import (
	errors_core "backend/internal/core/errors"
	"backend/internal/core/repository/postgres/pool"
	"context"
	"fmt"
)

func (r *UserRepository) DeleteUserTx(ctx context.Context, tx pool.Tx, id int) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	cmdTag, err := tx.Exec(ctx, `DELETE FROM users.users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("user with id %d: %w", id, errors_core.ErrNotFound)
	}

	return nil
}
