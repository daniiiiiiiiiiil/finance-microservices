package postgres

import (
	"fmt"

	"golang.org/x/net/context"
)

func (r *ShoppingRepository) GetTotalShopping(ctx context.Context, userID int) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var total int
	queryTotal := `
	SELECT COUNT(*) FROM shopping.shopping
	WHERE user_id = $1
`
	err := r.pool.QueryRow(ctx, queryTotal, userID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("GetShopping: could not get shopping list: %w", err)
	}
	return total, nil
}
