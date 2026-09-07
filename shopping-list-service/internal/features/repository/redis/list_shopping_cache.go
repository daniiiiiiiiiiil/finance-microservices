package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/cache"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/domain"
)

type ShoppingListCache struct {
	client *cache.RedisClient
}

func NewShoppingListCache(client *cache.RedisClient) *ShoppingListCache {
	return &ShoppingListCache{
		client: client,
	}
}

func (c *ShoppingListCache) GetShoppingList(ctx context.Context, limit, offset int) ([]domain.Shopping, bool) {
	key := fmt.Sprintf("shopping:%d:%d", limit, offset)
	var shoppingList []domain.Shopping
	if err := c.client.Get(ctx, key, &shoppingList); err != nil {
		return nil, false
	}
	return shoppingList, true
}

func (c *ShoppingListCache) SetShoppingList(ctx context.Context, shoppingList []domain.Shopping, limit, offset int) error {
	key := fmt.Sprintf("shopping:%d:%d", limit, offset)
	if err := c.client.Set(ctx, key, shoppingList, 5*time.Minute); err != nil {
		return fmt.Errorf("error setting shoppings: %w", err)
	}
	return nil
}

func (c *ShoppingListCache) InvalidateAllShoppingList(ctx context.Context) error {
	pattern := "shopping:list:*"
	var cursor uint64
	var keys []string

	for {
		var err error
		var batch []string
		batch, cursor, err = c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("error getting shoppings: %w", err)
		}
		keys = append(keys, batch...)
		if cursor == 0 {
			break
		}
	}
	if len(keys) > 0 {
		for _, key := range keys {
			if err := c.client.Delete(ctx, key); err != nil {
				return fmt.Errorf("error deleting shoppings: %w", err)
			}
		}
	}
	return nil
}
