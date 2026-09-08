package postgres

import (
	"context"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/repository/postgres/pool"
)

func (r *ShoppingRepository) DeleteShopping(ctx context.Context, tx pool.Tx, id int, userID int) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	sqlQuery := `DELETE FROM shopping.shopping WHERE id = $1 AND user_id = $2;`
	result, err := tx.Exec(ctx, sqlQuery, id, userID)
	if err != nil {
		return fmt.Errorf("delete shopping: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("shopping with id %d not found", id)
	}
	return nil
}
