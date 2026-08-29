package service

import (
	"fmt"
	"time"

	"context"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/features/finance/repository/postgres"
	"go.uber.org/zap"
)

func (s *FinanceService) DeleteUserTransactions(ctx context.Context, userID int) (int, error) {
	key := fmt.Sprintf("deleted:user_transactions:%d", userID)
	exists, err := s.redis.Exists(ctx, key)
	if err != nil {
		s.logger.Warn("failed to check if user transactions already deleted", zap.Error(err))
	}
	if exists > 0 {
		return 0, nil
	}

	count, err := s.repo.DeleteUserTransactions(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("error deleting user transactions: %w", err)
	}

	if err := s.redis.Set(ctx, key, true, 24*time.Hour); err != nil {
		s.logger.Warn("failed to mark user transactions as deleted", zap.Error(err))
	}

	go s.invalidateCache(context.Background(), userID)

	go s.sendUserTransactionsDeletedEvent(context.Background(), userID, count)

	return count, nil
}

func (s *FinanceService) GetMetrics(ctx context.Context) (postgres.Metrics, error) {
	key := "finance:metrics"
	var metrics postgres.Metrics
	err := s.redis.Get(ctx, key, &metrics)
	if err == nil {
		return metrics, nil
	}

	metrics, err = s.repo.GetMetrics(ctx)
	if err != nil {
		return postgres.Metrics{}, fmt.Errorf("get metrics: %w", err)
	}

	if err := s.redis.Set(ctx, key, metrics, 1*time.Minute); err != nil {
		s.logger.Warn("failed to cache finance metrics", zap.Error(err))
	}

	return metrics, nil
}
