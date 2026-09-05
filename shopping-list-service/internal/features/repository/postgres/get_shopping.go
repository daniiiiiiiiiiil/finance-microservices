package postgres

import (
	"errors"
	"fmt"

	"context"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/repository/postgres/pool"
	errors_my "github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/pkg/errors"
)

func (r *ShoppingRepository) GetShopping(ctx context.Context, id int) (domain.Shopping, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()
	sqlQuery := `
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
	WHERE id = $1
`
	var shopping ShoppingModel
	err := r.pool.QueryRow(ctx, sqlQuery, id).Scan(
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
		&shopping.CompletionDate,
	)
	if err != nil {
		if errors.Is(err, pool.ErrNoRows) {
			return domain.Shopping{}, fmt.Errorf("shopping with id %d:%w", id, errors_my.ErrNotFound)
		}
		return domain.Shopping{}, fmt.Errorf("get transaction: %w", err)
	}
	return shoppingDomainFromModel(shopping), nil
}
