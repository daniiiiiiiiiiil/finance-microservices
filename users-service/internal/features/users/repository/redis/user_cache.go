package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/cache"
	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/domain"
)

type UserCache struct {
	client *cache.RedisClient
}

func NewUserCache(client *cache.RedisClient) *UserCache {
	return &UserCache{client: client}
}

func (c *UserCache) GetUser(ctx context.Context, userID int) (domain.User, error) {
	key := fmt.Sprintf("user:%d", userID)
	var user domain.User
	if err := c.client.Get(ctx, key, &user); err != nil {
		return domain.User{}, fmt.Errorf("user not found in cache: %w", err)
	}
	return user, nil
}

func (c *UserCache) SetUser(ctx context.Context, user domain.User) error {
	key := fmt.Sprintf("user:%d", user.ID)
	if err := c.client.Set(ctx, key, user, 10*time.Minute); err != nil {
		return fmt.Errorf("failed to cache user: %w", err)
	}
	return nil
}

func (c *UserCache) DeleteUser(ctx context.Context, userID int) error {
	key := fmt.Sprintf("user:%d", userID)
	if err := c.client.Delete(ctx, key); err != nil {
		return fmt.Errorf("failed to delete user from cache: %w", err)
	}
	return nil
}

func (c *UserCache) InvalidateUser(ctx context.Context, userID int) error {
	return c.DeleteUser(ctx, userID)
}
