package redis

import (
	"fmt"
	"time"

	"context"
	"github.com/daniiiiiiiiiiil/finance-microservices/currency-service/internal/core/cache"
	"github.com/daniiiiiiiiiiil/finance-microservices/currency-service/internal/core/domain"
)

type RateCache struct {
	client *cache.RedisClient
}

func NewRateCache(client *cache.RedisClient) *RateCache {
	return &RateCache{
		client: client,
	}
}

func (c *RateCache) GetRate(ctx context.Context, base string) (*domain.Rate, error) {
	key := fmt.Sprintf("currency:rate:%s", base)
	var rate domain.Rate
	if err := c.client.Get(ctx, key, &rate); err != nil {
		return nil, fmt.Errorf("get rate from redis: %w", err)
	}
	return &rate, nil
}

func (c *RateCache) SetRate(ctx context.Context, rate domain.Rate, ttl time.Duration) error {
	key := fmt.Sprintf("currency:rate:%s", rate.Base)
	if err := c.client.Set(ctx, key, &rate, ttl); err != nil {
		return fmt.Errorf("set rate to redis: %w", err)
	}
	return nil
}

func (c *RateCache) GetConvertedUSD(ctx context.Context, txID int) (float64, error) {
	key := fmt.Sprintf("currency:tx:%d:usd", txID)
	var converted float64
	if err := c.client.Get(ctx, key, &converted); err != nil {
		return 0, fmt.Errorf("get rate from redis: %w", err)
	}
	return converted, nil
}

func (c *RateCache) SetConvertedUSD(ctx context.Context, txID int, amount float64, ttl time.Duration) error {
	key := fmt.Sprintf("currency:tx:%d:usd", txID)
	if err := c.client.Set(ctx, key, amount, ttl); err != nil {
		return fmt.Errorf("set rate to redis: %w", err)
	}
	return nil
}

func (c *RateCache) DeleteRate(ctx context.Context, base string) error {
	key := fmt.Sprintf("currency:rate:%s", base)
	if err := c.client.Delete(ctx, key); err != nil {
		return fmt.Errorf("delete rate from redis: %w", err)
	}
	return nil
}

func (c *RateCache) DeleteConvertedUSD(ctx context.Context, txID int) error {
	key := fmt.Sprintf("currency:tx:%d:usd", txID)
	return c.client.Delete(ctx, key)
}

func (c *RateCache) Exists(ctx context.Context, key string) (int64, error) {
	return c.client.Exists(ctx, key)
}

func (c *RateCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl)
}

func (c *RateCache) SetTransactionUSD(ctx context.Context, tx domain.TransactionUSD, ttl time.Duration) error {
	key := fmt.Sprintf("currency:tx:%d:usd", tx.TransactionID)
	return c.client.Set(ctx, key, tx, ttl)
}

func (c *RateCache) GetTransactionUSD(ctx context.Context, txID int) (domain.TransactionUSD, error) {
	key := fmt.Sprintf("currency:tx:%d:usd", txID)
	var tx domain.TransactionUSD
	if err := c.client.Get(ctx, key, &tx); err != nil {
		return domain.TransactionUSD{}, fmt.Errorf("get rate from redis: %w", err)
	}
	return tx, nil
}
