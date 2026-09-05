package postgres

import (
	"database/sql"
	"errors"
	"fmt"

	"context"
)

func (r *ShoppingRepository) CompletedShopping(ctx context.Context, id int, completed bool) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	sqlQuery := `
	UPDATE shopping.shopping 
	SET	completed = $1
	WHERE id = $2
	RETURNING id,completed;
`
	err := r.pool.QueryRow(ctx, sqlQuery, completed, id).Scan(&id, completed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("shopping with id %d not found", id)
		}
		return fmt.Errorf("shopping with id %d error %w", id, err)
	}
	return nil
}
