package postgres

import (
	"fmt"
	"time"

	"context"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/repository/postgres/pool"
)

func (r *ShoppingRepository) CreateShopping(ctx context.Context, tx pool.Tx, shopping domain.Shopping, userID int) (domain.Shopping, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	    INSERT INTO shopping.shopping (
	                				   user_id,
	                                   title,
	                                   description,
	                                   amount_now,
	                                   amount_finish,
	                                   image_key,
	                                   completed,
	                                   created_at,
	                                   completion_date)
	    VALUES ($1, $2, $3, $4, $5, $6, $7,$8,$9)
	     RETURNING id, version, title, description, amount_now, amount_finish,
	        image_key, completed, created_at, updated_at, completed_at, completion_date
`
	var model ShoppingModel
	err := tx.QueryRow(ctx, query,
		userID,
		shopping.Title,
		shopping.Description,
		shopping.AmountNow,
		shopping.AmountFinish,
		shopping.ImageKey,
		shopping.Completed,
		time.Now(),
		shopping.CompletionDate).Scan(
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
		&model.CompletedAt,
		&model.CompletionDate,
	)
	if err != nil {
		return domain.Shopping{}, fmt.Errorf("error inserting shopping: %w", err)
	}
	return shoppingDomainFromModel(model), nil
}
