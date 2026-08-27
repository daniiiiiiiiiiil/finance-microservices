package ports

import (
	"time"

	"golang.org/x/net/context"
)

type RedisClientInterface interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
	ExistsBool(ctx context.Context, key string) (bool, error)
	Incr(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
}

type RateLimitInterface interface {
	Check(ctx context.Context, key string, limit int64, ttl time.Duration) (bool, error)
	Reset(ctx context.Context, key string) error
}

type BlacklistInterface interface {
	Add(ctx context.Context, token string, ttl time.Duration) error
	IsBlacklisted(ctx context.Context, token string) (bool, error)
	Remove(ctx context.Context, token string) error
}
