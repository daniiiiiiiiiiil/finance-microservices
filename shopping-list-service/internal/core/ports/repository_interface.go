package ports

import (
	"context"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/repository/postgres/pool"
)

type ShoppingListRepository interface {
	CreateShopping(ctx context.Context, tx pool.Tx, shopping domain.Shopping, userID int) (domain.Shopping, error)
	ListShopping(ctx context.Context, userID, limit, offset int) ([]domain.Shopping, int, error)
	GetShopping(ctx context.Context, id int, userID int) (domain.Shopping, error)
	UpdateShopping(ctx context.Context, shopping *domain.Shopping, userID int) (domain.Shopping, error)
	DeleteShopping(ctx context.Context, id int, userID int) error
	CompletedShopping(ctx context.Context, id int, userID int, completed bool) error
	GetTotalShopping(ctx context.Context, userID int) (int, error)
}
