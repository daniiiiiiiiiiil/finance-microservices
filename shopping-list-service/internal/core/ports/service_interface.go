package ports

import (
	"context"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/domain"
)

type ShoppingServiceInterface interface {
	CreateShopping(ctx context.Context, shopping domain.Shopping, userID int) (domain.Shopping, error)
	GetShopping(ctx context.Context, id int, userID int) (domain.Shopping, error)
	GetTotal(ctx context.Context, userID int) (int, error)
	ListShopping(ctx context.Context, limit, offset, userID int) ([]domain.Shopping, int, error)
	CompletedShopping(ctx context.Context, id int, completed bool, userID int) error
	DeleteShoppingList(ctx context.Context, id int, userID int) error
	UpdateShopping(ctx context.Context, shopping *domain.Shopping, userID int) (domain.Shopping, error)
}
