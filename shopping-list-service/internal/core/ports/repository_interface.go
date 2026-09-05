package ports

import (
	"context"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/repository/postgres/pool"
)

type ShoppingListRepository interface {
	CreateShopping(ctx context.Context, tx pool.Tx, shopping domain.Shopping) (domain.Shopping, error)
	ListShopping(ctx context.Context, limit, offset int) ([]domain.Shopping, int, error)
	GetShopping(ctx context.Context, id int) (domain.Shopping, error)
	UpdateShopping(ctx context.Context, shopping *domain.Shopping) (domain.Shopping, error)
	DeleteShopping(ctx context.Context, id int) error
	CompletedShopping(ctx context.Context, id int, completed bool) error
}
