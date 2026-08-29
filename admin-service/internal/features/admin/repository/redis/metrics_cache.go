package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/internal/core/cache"
	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/internal/core/ports"
)

type MetricsCache struct {
	client *cache.RedisClient
}

func NewMetricsCache(client *cache.RedisClient) *MetricsCache {
	return &MetricsCache{
		client: client,
	}
}

func (c *MetricsCache) Get(ctx context.Context) (ports.Metrics, error) {
	key := "admin:metrics"
	var metrics ports.Metrics
	if err := c.client.Get(ctx, key, &metrics); err != nil {
		return ports.Metrics{}, fmt.Errorf("Failed to get metrics from redis: %w", err)
	}
	return metrics, nil
}

func (c *MetricsCache) Set(ctx context.Context, metrics ports.Metrics, ttl time.Duration) error {
	key := "admin:metrics"
	if err := c.client.Set(ctx, key, metrics, ttl); err != nil {
		return fmt.Errorf("Failed to set metrics to redis: %w", err)
	}
	return nil
}

func (c *MetricsCache) Delete(ctx context.Context) error {
	key := "admin:metrics"
	if err := c.client.Delete(ctx, key); err != nil {
		return fmt.Errorf("Failed to delete metrics from redis: %w", err)
	}
	return nil
}
