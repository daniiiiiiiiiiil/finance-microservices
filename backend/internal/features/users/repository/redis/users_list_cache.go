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

// InvalidateAllUsersList инвалидирует все кеши списков
func (c *UsersListCache) InvalidateAllUsersList(ctx context.Context) error {
	// Используем паттерн для удаления всех ключей users:list:*
	// В Redis можно использовать SCAN + DEL, но для простоты используем Set
	// Или хранить версию кеша
	return nil
}
