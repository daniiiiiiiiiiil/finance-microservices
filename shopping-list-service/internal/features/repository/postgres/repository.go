package postgres

import "github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/repository/postgres/pool"

type ShoppingRepository struct {
	pool pool.Pool
}

func NewShoppingRepository(pool pool.Pool) *ShoppingRepository {
	return &ShoppingRepository{pool: pool}
}
