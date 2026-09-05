package postgres

import (
	"fmt"

	"context"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/domain"
)

func (r *ShoppingRepository) ListShopping(ctx context.Context, limit, offset int) ([]domain.Shopping, int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var total int
	queryTotal := `
	SELECT COUNT(*) FROM shopping.shopping
`
	err := r.pool.QueryRow(ctx, queryTotal).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("GetShopping: could not get shopping list: %w", err)
	}
	querySQL := `
	SELECT 
		id,
		version,
		title, 
		description,
		amount_now,
		amount_finish,
		image_key,
		completed,
		created_at,
		updated_at,
		completed_at,
		completion_date
	FROM shopping.shopping
	ORDER BY id
	LIMIT $1 OFFSET $2
`
	rows, err := r.pool.Query(ctx, querySQL, limit, offset)
	if err != nil {
		return nil, total, fmt.Errorf("GetShopping: could not get shopping list: %w", err)
	}
	defer rows.Close()
	var shoppings []domain.Shopping
	for rows.Next() {
		var shopping domain.Shopping
		err := rows.Scan(
			&shopping.ID,
			&shopping.Version,
			&shopping.Title,
			&shopping.Description,
			&shopping.AmountNow,
			&shopping.AmountFinish,
			&shopping.ImageKey,
			&shopping.Completed,
			&shopping.CreatedAt,
			&shopping.UpdatedAt,
			&shopping.CompletedAt,
			&shopping.CompletionDate)
		if err != nil {
			return nil, total, fmt.Errorf("GetShopping: could not scan shopping row: %w", err)
		}
		shoppings = append(shoppings, shopping)
	}
	if err := rows.Err(); err != nil {
		return nil, total, fmt.Errorf("GetShopping: could not get shopping list: %w", err)
	}
	return shoppings, total, nil
}
