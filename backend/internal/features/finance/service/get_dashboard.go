package service

import (
	"backend/internal/core/domain"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func (s *FinanceService) GetDashboard(ctx context.Context, userID int) (domain.Dashboard, error) {
	key := fmt.Sprintf("dashboard:%d", userID)
	var dashboard domain.Dashboard

	err := s.redis.Get(ctx, key, &dashboard)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			dashboard, err = s.repo.GetDashboard(ctx, userID)
			if err != nil {
				return domain.Dashboard{}, fmt.Errorf("Failed to get dashboard: %w", err)
			}
			if err := s.redis.Set(ctx, key, dashboard, 10*time.Minute); err != nil {
				fmt.Printf("failed to cache dashboard: %v\n", err)
			}
			return dashboard, nil
		}
		return domain.Dashboard{}, fmt.Errorf("Failed to get dashboard: %w", err)
	}
	return dashboard, nil

}
