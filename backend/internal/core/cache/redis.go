package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisInterface interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Incr(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
}

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(addr string) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisClient{
		client: client,
	}
}

var _ RedisInterface = (*RedisClient)(nil)

func (r *RedisClient) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := r.client.Get(ctx, key).Result() //получаем строку из редиса
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(val), dest); err != nil { //преобразуем в структуру go
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}
	return nil
}

func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value) // превращаем в json
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil { //сохраняем json по ключу и даем ему срок жизни
		return fmt.Errorf("failed to set value: %w", err)
	}
	return nil
}

func (r *RedisClient) Delete(ctx context.Context, key string) error { // просто удаляет ключ
	return r.client.Del(ctx, key).Err()
}

func (r *RedisClient) Incr(ctx context.Context, key string) (int64, error) { // увеличивает число на 1 если не существует создает со значением 1
	return r.client.Incr(ctx, key).Result()
}

func (r *RedisClient) Expire(ctx context.Context, key string, ttl time.Duration) error { //устанавливаем время через сколько ключ будет удален но ключ уже должен существовать иначе ничего не будет
	return r.client.Expire(ctx, key, ttl).Err()
}
