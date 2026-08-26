package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/cache"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/domain"
)

type DashboardCache struct {
	client *cache.RedisClient
}

func NewDashboardCache(client *cache.RedisClient) *DashboardCache {
	return &DashboardCache{client: client}
}

func (c *DashboardCache) Get(ctx context.Context, userID int) (domain.Dashboard, error) {
	key := fmt.Sprintf("dashboard:%d", userID)
	var dashboard domain.Dashboard
	if err := c.client.Get(ctx, key, &dashboard); err != nil {
		return domain.Dashboard{}, fmt.Errorf("Failed to get dashboard from redis: %w", err)
	}
	return dashboard, nil
}

func (c *DashboardCache) Set(ctx context.Context, userID int, dashboard domain.Dashboard) error {
	key := fmt.Sprintf("dashboard:%d", userID)
	if err := c.client.Set(ctx, key, dashboard, time.Hour*24); err != nil {
		return fmt.Errorf("Failed to set dashboard to redis: %w", err)
	}
	return nil
}

func (c *DashboardCache) Delete(ctx context.Context, userID int) error {
	key := fmt.Sprintf("dashboard:%d", userID)
	if err := c.client.Delete(ctx, key); err != nil {
		return fmt.Errorf("Failed to delete dashboard from redis: %w", err)
	}
	return nil
}
