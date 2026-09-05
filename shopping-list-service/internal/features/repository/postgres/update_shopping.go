package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"context"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/domain"
)

func (r *ShoppingRepository) UpdateShopping(ctx context.Context, shopping *domain.Shopping) (domain.Shopping, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	sqlQuery := `
	UPDATE shopping.shopping SET 
	            version = version + 1,
				title = $1,
				description = $2,
				amount_now = $3,
				amount_finish = $4,
				image_key = $5,
				completed = $6,
				updated_at = $7,
				completion_date = $8
	WHERE id = $9 AND version = $10
	RETURNING id, version, title, description, amount_now, anount_finish,
	image_key,completed,created_at, update_at,completion_date
`
	var model ShoppingModel
	err := r.pool.QueryRow(ctx, sqlQuery,
		shopping.Title,
		shopping.Description,
		shopping.AmountNow,
		shopping.AmountFinish,
		shopping.ImageKey,
		shopping.Completed,
		time.Now(),
		shopping.CompletionDate,
		shopping.ID,
		shopping.Version).Scan(
		&model.ID,
		&model.Version,
		&model.Title,
		&model.Description,
		&model.AmountNow,
		&model.AmountFinish,
		&model.ImageKey,
		&model.Completed,
		&model.CreatedAt,
		&model.UpdatedAt,
		&model.CompletionDate,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Shopping{}, fmt.Errorf("shopping with id %d not found", shopping.ID)
		}
		return domain.Shopping{}, fmt.Errorf("shopping with id %d:%w", shopping.ID, err)
	}
	return shoppingDomainFromModel(model), nil
}
