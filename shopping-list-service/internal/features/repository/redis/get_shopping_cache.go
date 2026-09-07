package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/cache"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/domain"
)

type ShoppingCache struct {
	client *cache.RedisClient
}

func NewShoppingCache(client *cache.RedisClient) *ShoppingCache {
	return &ShoppingCache{
		client: client,
	}
}

func (c *ShoppingCache) GetShopping(ctx context.Context, id int) (domain.Shopping, error) {
	key := fmt.Sprintf("shopping:%d", id)
	var shopping domain.Shopping
	if err := c.client.Get(ctx, key, &shopping); err != nil {
		return domain.Shopping{}, fmt.Errorf("error getting shopping from redis: %w", err)
	}
	return shopping, nil
}

func (c *ShoppingCache) SetShopping(ctx context.Context, shopping domain.Shopping) error {
	key := fmt.Sprintf("shopping:%d", shopping.ID)
	if err := c.client.Set(ctx, key, shopping, 10*time.Minute); err != nil {
		return fmt.Errorf("error setting shopping in redis: %w", err)
	}
	return nil
}

func (c *ShoppingCache) DeleteShopping(ctx context.Context, id int) error {
	key := fmt.Sprintf("shopping:%d", id)
	if err := c.client.Delete(ctx, key); err != nil {
		return fmt.Errorf("error deleting shopping from redis: %w", err)
	}
	return nil
}

func (c *ShoppingCache) InvalidateUser(ctx context.Context, id int) error {
	return c.DeleteShopping(ctx, id)
}
