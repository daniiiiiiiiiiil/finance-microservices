package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/internal/core/cache"
	service_admin "github.com/daniiiiiiiiiiil/finance-microservices/admin-service/internal/features/admin/service"
)

var _ MetricsCacheInterface = (*MetricsCache)(nil)

type MetricsCache struct {
	client *cache.RedisClient
}

func NewMetricsCache(client *cache.RedisClient) *MetricsCache {
	return &MetricsCache{
		client: client,
	}
}

func (c *MetricsCache) Get(ctx context.Context) (service_admin.Metrics, error) {
	key := "admin:metrics"
	var metrics service_admin.Metrics
	if err := c.client.Get(ctx, key, &metrics); err != nil {
		return service_admin.Metrics{}, fmt.Errorf("Failed to get metrics from redis: %w", err)
	}
	return metrics, nil
}

func (c *MetricsCache) Set(ctx context.Context, metrics service_admin.Metrics, ttl time.Duration) error {
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
