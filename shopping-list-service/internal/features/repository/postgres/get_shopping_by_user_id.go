package postgres

import (
	"context"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/domain"
)

func (r *ShoppingRepository) GetShoppingsByUserID(ctx context.Context, userID int) ([]domain.Shopping, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT 
		id,
		version,
		user_id,
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
	WHERE user_id = $1
	ORDER BY id
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get shoppings by user: %w", err)
	}
	defer rows.Close()

	var shoppings []domain.Shopping
	for rows.Next() {
		var model ShoppingModel
		err := rows.Scan(
			&model.ID,
			&model.Version,
			&model.UserID,
			&model.Title,
			&model.Description,
			&model.AmountNow,
			&model.AmountFinish,
			&model.ImageKey,
			&model.Completed,
			&model.CreatedAt,
			&model.UpdatedAt,
			&model.CompletedAt,
			&model.CompletionDate,
		)
		if err != nil {
			return nil, fmt.Errorf("scan shopping: %w", err)
		}
		shoppings = append(shoppings, shoppingDomainFromModel(model))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return shoppings, nil
}
