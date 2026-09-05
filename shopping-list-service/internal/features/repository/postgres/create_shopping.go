package postgres

import (
	"fmt"
	"time"

	"context"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/repository/postgres/pool"
)

func (r *ShoppingRepository) CreateShopping(ctx context.Context, tx pool.Tx, shopping domain.Shopping) (domain.Shopping, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.pool.OpTimeout())
	defer cancel()

	query := `
	    INSERT INTO shopping.shopping (
	                                   title,
	                                   description,
	                                   amount_now,
	                                   amount_finish,
	                                   image_key,
	                                   completed,
	                                   created_at,
	                                   completion_date)
	    VALUES ($1, $2, $3, $4, $5, $6, $7,$8)
`
	var model ShoppingModel
	err := tx.QueryRow(ctx, query,
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
