package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func (r *ShoppingRepository) DeleteShopping(ctx context.Context, id int) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	sqlQuery := `DELETE FROM shopping.shopping WHERE id = $1;`
	err := r.pool.QueryRow(ctx, sqlQuery, id).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("shopping with id %d not found", id)
		}
		return fmt.Errorf("shopping with id %d:%w", id, err)
	}
	return nil
}
