package service

import (
	service "backend/internal/features/admin/service"
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

func (s *FinanceService) GetMetrics(ctx context.Context) (service.Metrics, error) {
	key := "finance:metrics"
	var metrics service.Metrics
	err := s.redis.Get(ctx, key, &metrics)
	if err == nil {
		return metrics, nil
	}

	metrics, err = s.repo.GetMetrics(ctx)
	if err != nil {
		return service.Metrics{}, fmt.Errorf("get metrics: %w", err)
	}

	if err := s.redis.Set(ctx, key, metrics, 1*time.Minute); err != nil {
		s.logger.Warn("failed to cache finance metrics", zap.Error(err))
	}

	return metrics, nil
}
