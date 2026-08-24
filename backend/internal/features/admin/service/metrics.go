package service_admin

import (
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

func (s *AdminService) GetMetrics(ctx context.Context) (Metrics, error) {
	key := "admin:metrics"
	var metrics Metrics

	err := s.redis.Get(ctx, key, &metrics)
	if err == nil {
		s.logger.Debug("metrics served from redis",
			zap.Int("total_users", metrics.TotalUsers),
			zap.Int("total_transactions", metrics.TotalTransactions),
			zap.Float64("total_balance", metrics.TotalBalance),
		)
		return metrics, nil
	}

	if errors.Is(err, redis.Nil) {
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			auth := md.Get("authorization")
			if len(auth) > 0 {
				ctx = metadata.AppendToOutgoingContext(ctx, "authorization", auth[0])
			}
		}
		usersMetrics, err := s.userClient.GetMetrics(ctx)
		if err != nil {
			return Metrics{}, fmt.Errorf("get users metrics: %w", err)
		}
		metrics.TotalUsers = usersMetrics.TotalUsers

		financeMetrics, err := s.financeClient.GetMetrics(ctx)
		if err != nil {
			return Metrics{}, fmt.Errorf("get finance metrics: %w", err)
		}
		metrics.TotalTransactions = financeMetrics.TotalTransactions
		metrics.TotalBalance = financeMetrics.TotalBalance

		if err := s.redis.Set(ctx, key, metrics, 10*time.Minute); err != nil {
			s.logger.Warn("failed to cache metrics in redis", zap.Error(err))
			return Metrics{}, fmt.Errorf("set metrics to redis: %w", err)
		}

		return metrics, nil
	}

	s.logger.Error("failed to get metrics from redis", zap.Error(err))
	return Metrics{}, fmt.Errorf("get metrics from redis: %w", err)
}
