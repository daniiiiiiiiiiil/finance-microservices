package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/core/cache"
)

type BlacklistCache struct {
	client *cache.RedisClient
}

func NewBlacklistCache(client *cache.RedisClient) *BlacklistCache {
	if client == nil {
		panic("redis client is nil in NewBlacklistCache")
	}
	return &BlacklistCache{client: client}
}

func (c *BlacklistCache) Add(ctx context.Context, token string, ttl time.Duration) error {
	if c.client == nil || c.client.GetClient() == nil {
		return fmt.Errorf("redis client is nil")
	}
	key := fmt.Sprintf("blacklist:%s", token)
	if err := c.client.Set(ctx, key, "blacklisted", ttl); err != nil {
		return fmt.Errorf("failed to add token to blacklist: %w", err)
	}
	return nil
}

func (c *BlacklistCache) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	if c.client == nil || c.client.GetClient() == nil {
		return false, nil
	}
	key := fmt.Sprintf("blacklist:%s", token)
	var val string
	err := c.client.Get(ctx, key, &val)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (c *BlacklistCache) Remove(ctx context.Context, token string) error {
	if c.client == nil || c.client.GetClient() == nil {
		return nil
	}
	key := fmt.Sprintf("blacklist:%s", token)
	return c.client.Delete(ctx, key)
}
