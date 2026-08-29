package ports

import (
	"time"

	"context"
	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/domain"
)

type RedisInterface interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Incr(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	Exists(ctx context.Context, key string) (int64, error)
}

type UserCacheInterface interface {
	GetUser(ctx context.Context, userID int) (domain.User, error)
	SetUser(ctx context.Context, user domain.User) error
	DeleteUser(ctx context.Context, userID int) error
	InvalidateUser(ctx context.Context, userID int) error
}

type UsersListCacheInterface interface {
	GetUsersList(ctx context.Context, limit, offset int) ([]domain.User, bool)
	SetUsersList(ctx context.Context, users []domain.User, limit, offset int) error
	InvalidateAllUsersList(ctx context.Context) error
}
