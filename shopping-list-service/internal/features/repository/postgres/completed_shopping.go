package postgres

import (
	"database/sql"
	"errors"
	"fmt"

	"context"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/repository/postgres/pool"
)

func (r *ShoppingRepository) CompletedShopping(ctx context.Context, tx pool.Tx, id, userID int, completed bool) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	sqlQuery := `
	UPDATE shopping.shopping 
	SET	completed = $1
	WHERE id = $2 AND user_id = $3
	RETURNING id,completed;
`
	var returnedID int
	var returnedCompleted bool
	err := tx.QueryRow(ctx, sqlQuery, completed, id, userID).Scan(&returnedID, &returnedCompleted)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("shopping with id %d not found", id)
		}
		return fmt.Errorf("shopping with id %d error %w", id, err)
	}
	return nil
}
