package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(addr string) *RedisClient {
	return &RedisClient{client: redis.NewClient(&redis.Options{Addr: addr})}
}

func (r *RedisClient) Get(ctx context.Context, key string, dest any) error {
	value, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(value), dest); err != nil {
		return fmt.Errorf("unmarshal value: %w", err)
	}
	return nil
}

func (r *RedisClient) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal value: %w", err)
	}
	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("set value: %w", err)
	}
	return nil
}

func (r *RedisClient) SetNX(ctx context.Context, key string, value any, ttl time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("marshal value: %w", err)
	}
	return r.client.SetNX(ctx, key, data, ttl).Result()
}

func (r *RedisClient) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisClient) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

func (r *RedisClient) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return r.client.Expire(ctx, key, ttl).Err()
}

func (r *RedisClient) Exists(ctx context.Context, key string) (int64, error) {
	return r.client.Exists(ctx, key).Result()
}

func (r *RedisClient) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}
