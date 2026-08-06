package redis

import (
	"backend/internal/core/cache"
	"context"
	"fmt"
	"time"
)

type CategoriesCache struct {
	client *cache.RedisClient
}

func NewCategoriesCache(client *cache.RedisClient) *CategoriesCache {
	return &CategoriesCache{client: client}
}

func (c *CategoriesCache) Get(ctx context.Context, userID int) ([]string, error) {
	key := fmt.Sprintf("categories:%d", userID)
	var categories []string
	err := c.client.Get(ctx, key, &categories)
	if err != nil {
		return nil, fmt.Errorf("Failed to get categories from redis: %w", err)
	}
	return categories, nil
}

func (c *CategoriesCache) Set(ctx context.Context, userID int, categories []string) error {
	key := fmt.Sprintf("categories:%d", userID)

	if err := c.client.Set(ctx, key, categories, 10*time.Minute); err != nil {
		return fmt.Errorf("Failed to set categories in redis: %w", err)
	}
	return nil
}

func (c *CategoriesCache) Delete(ctx context.Context, userID int) error {
	key := fmt.Sprintf("categories:%d", userID)
	if err := c.client.Delete(ctx, key); err != nil {
		return fmt.Errorf("Failed to delete categories from redis: %w", err)
	}
	return nil
}
