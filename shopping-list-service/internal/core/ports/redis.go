package ports

import (
	"context"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/domain"
)

type RedisInterface interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Incr(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	Exists(ctx context.Context, key string) (int64, error)
}

type ShoppingCacheInterface interface {
	GetShopping(ctx context.Context, id int) (domain.Shopping, error)
	SetShopping(ctx context.Context, shopping domain.Shopping) error
	DeleteShopping(ctx context.Context, id int) error
	InvalidateUser(ctx context.Context, id int) error
}

type ShoppingListCacheInterface interface {
	GetShoppingList(ctx context.Context, limit, offset int) ([]domain.Shopping, bool)
	SetShoppingList(ctx context.Context, shoppingList []domain.Shopping, limit, offset int) error
	InvalidateAllShoppingList(ctx context.Context) error
}
