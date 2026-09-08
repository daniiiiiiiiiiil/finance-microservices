package ports

import (
	"context"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/repository/postgres/pool"
)

type ShoppingListRepository interface {
	CreateShopping(ctx context.Context, tx pool.Tx, shopping domain.Shopping, userID int) (domain.Shopping, error)
	GetShopping(ctx context.Context, id, userID int) (domain.Shopping, error)
	ListShopping(ctx context.Context, tx pool.Tx, userID, limit, offset int) ([]domain.Shopping, int, error)
	UpdateShopping(ctx context.Context, tx pool.Tx, shopping *domain.Shopping, userID int) (domain.Shopping, error)
	DeleteShopping(ctx context.Context, tx pool.Tx, id int, userID int) error
	CompletedShopping(ctx context.Context, tx pool.Tx, id, userID int, completed bool) error
	GetTotalShopping(ctx context.Context, tx pool.Tx, userID int) (int, error)
}
