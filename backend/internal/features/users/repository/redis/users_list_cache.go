// internal/features/users/repository/redis/users_list_cache.go
package redis

import (
	"backend/internal/core/cache"
	"backend/internal/core/domain"
	"context"
	"fmt"
	"time"
)

type UsersListCache struct {
	client *cache.RedisClient
}

func NewUsersListCache(client *cache.RedisClient) *UsersListCache {
	return &UsersListCache{client: client}
}

func (c *UsersListCache) GetUsersList(ctx context.Context, limit, offset int) ([]domain.User, bool) {
	key := fmt.Sprintf("users:list:%d:%d", limit, offset)
	var users []domain.User
	if err := c.client.Get(ctx, key, &users); err != nil {
		return nil, false
	}
	return users, true
}

func (c *UsersListCache) SetUsersList(ctx context.Context, users []domain.User, limit, offset int) error {
	key := fmt.Sprintf("users:list:%d:%d", limit, offset)
	if err := c.client.Set(ctx, key, users, 5*time.Minute); err != nil {
		return fmt.Errorf("failed to cache users list: %w", err)
	}
	return nil
}

func (c *UsersListCache) InvalidateAllUsersList(ctx context.Context) error {
	pattern := "users:list:*"
	var cursor uint64
	var keys []string

	for {
		var err error
		var batch []string
		batch, cursor, err = c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("failed to scan keys: %w", err)
		}
		keys = append(keys, batch...)
		if cursor == 0 {
			break
		}
	}

	if len(keys) > 0 {
		for _, key := range keys {
			if err := c.client.Delete(ctx, key); err != nil {
				return fmt.Errorf("failed to delete key %s: %w", key, err)
			}
		}
	}

	return nil
}
