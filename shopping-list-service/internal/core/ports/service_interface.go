package ports

import (
	"context"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/domain"
)

type ShoppingServiceInterface interface {
	CreateShopping(ctx context.Context, shopping domain.Shopping) (domain.Shopping, error)
}
