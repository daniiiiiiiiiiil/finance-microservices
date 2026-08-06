package redis

import (
	"backend/internal/core/cache"
	"context"
	"fmt"
	"time"
)

type RateLimitCache struct {
	client *cache.RedisClient
}

func NewRateLimitCache(client *cache.RedisClient) *RateLimitCache {
	return &RateLimitCache{client: client}
}

func (c *RateLimitCache) Check(ctx context.Context, key string, limit int64, ttl time.Duration) (bool, error) {
	count, err := c.client.Incr(ctx, key)
	if err != nil {
		return false, fmt.Errorf("failed to increment rate limit: %w", err)
	}

	if count == 1 {
		if err := c.client.Expire(ctx, key, ttl); err != nil {
			return false, fmt.Errorf("failed to set expire: %w", err)
		}
	}

	return count <= limit, nil
}

func (c *RateLimitCache) Reset(ctx context.Context, key string) error {
	return c.client.Delete(ctx, key)
}
