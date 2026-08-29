package postgres

import (
	"database/sql"
	"fmt"

	"context"
	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/repository/postgres/pool"
)

func (r *UserRepository) UpdateStatusTx(ctx context.Context, tx pool.Tx, id int, status string) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `UPDATE users.users 
				SET status = $1, version = version + 1
                WHERE id = $2`

	cmdTag, err := tx.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("update status tx: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("update status tx: rows affected: %w", sql.ErrNoRows)
	}
	return nil
}
