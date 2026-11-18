package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/cache"
)

type DedupStore struct {
	client *cache.RedisClient
}

func NewDedupStore(client *cache.RedisClient) *DedupStore {
	return &DedupStore{client: client}
}

func (s *DedupStore) Claim(ctx context.Context, key string, ttlSeconds int) (bool, error) {
	ttl := time.Duration(ttlSeconds) * time.Second
	claimed, err := s.client.SetNX(ctx, key, time.Now().UTC(), ttl)
	if err != nil {
		return false, fmt.Errorf("claim key: %w", err)
	}
	return claimed, nil
}

func (s *DedupStore) Release(ctx context.Context, key string) error {
	if err := s.client.Delete(ctx, key); err != nil {
		return fmt.Errorf("release key: %w", err)
	}
	return nil
}
