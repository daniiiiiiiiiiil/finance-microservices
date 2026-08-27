package redis

import (
	"context"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/core/cache"
)

type RateLimitCache struct {
	client *cache.RedisClient
}

func NewRateLimitCache(client *cache.RedisClient) *RateLimitCache {
	if client == nil {
		panic("redis client is nil in NewRateLimitCache")
	}
	return &RateLimitCache{client: client}
}

func (c *RateLimitCache) Check(ctx context.Context, key string, limit int64, ttl time.Duration) (bool, error) {
	if c.client == nil {
		return true, nil
	}

	if c.client.GetClient() == nil {
		return true, nil
	}

	count, err := c.client.Incr(ctx, key)
	if err != nil {
		return true, nil
	}

	if count == 1 {
		if err := c.client.Expire(ctx, key, ttl); err != nil {
			return true, nil
		}
	}

	return count <= limit, nil
}

func (c *RateLimitCache) Reset(ctx context.Context, key string) error {
	if c.client == nil || c.client.GetClient() == nil {
		return nil
	}
	return c.client.Delete(ctx, key)
}
