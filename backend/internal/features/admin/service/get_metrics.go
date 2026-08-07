package service_admin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func (s *AdminService) GetMetrics(ctx context.Context) (Metrics, error) {
	key := "admin:metrics"
	var metrics Metrics

	err := s.redis.Get(ctx, key, &metrics)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			metrics, err = s.repo.GetMetrics(ctx)
			if err != nil {
				return Metrics{}, fmt.Errorf("get metrics from postgres: %w", err)
			}
			if err := s.redis.Set(ctx, key, metrics, 10*time.Minute); err != nil {
				return Metrics{}, fmt.Errorf("set metrics to redis: %w", err)
			}
			return metrics, nil
		}
		return Metrics{}, fmt.Errorf("get metrics from redis/postgres: %w", err)
	}
	return metrics, nil
}
