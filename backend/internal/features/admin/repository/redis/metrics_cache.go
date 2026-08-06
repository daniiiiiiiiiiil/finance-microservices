package redis

import (
	"backend/internal/core/cache"
	"backend/internal/features/admin/service"
	"context"
	"fmt"
	"time"
)

type MetricsCache struct {
	client *cache.RedisClient
}

func NewMetricsCache(client *cache.RedisClient) *MetricsCache {
	return &MetricsCache{
		client: client,
	}
}

func (c *MetricsCache) Get(ctx context.Context) (service.Metrics, error) {
	key := "admin:metrics"
	var metrics service.Metrics
	if err := c.client.Get(ctx, key, &metrics); err != nil {
		return service.Metrics{}, fmt.Errorf("Failed to get metrics from redis: %w", err)
	}
	return metrics, nil
}

func (c *MetricsCache) Set(ctx context.Context, metrics service.Metrics, ttl time.Duration) error {
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
