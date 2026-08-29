package service

import (
	"errors"
	"fmt"
	"time"

	"context"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/domain"
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

func (s *FinanceService) GetCategories(ctx context.Context, userID int) ([]string, error) {
	key := fmt.Sprintf("categories:%d", userID)
	var categories []string

	err := s.redis.Get(ctx, key, &categories)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			categories, err = s.repo.GetCategories(ctx, userID)
			if err != nil {
				return nil, fmt.Errorf("failed to get categories from DB: %w", err)
			}

			go func() {
				if err := s.redis.Set(context.Background(), key, categories, 24*time.Hour); err != nil {
					fmt.Printf("failed to cache categories: %v\n", err)
				}
			}()

			return categories, nil
		}

		fmt.Printf("redis get failed, falling back to DB: %v\n", err)
		categories, err = s.repo.GetCategories(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get categories from DB after redis error: %w", err)
		}
		return categories, nil
	}

	return categories, nil
}

func (s *FinanceService) invalidateCache(ctx context.Context, userID int) {
	dashboardKey := fmt.Sprintf("dashboard:%d", userID)
	if err := s.redis.Delete(ctx, dashboardKey); err != nil {
		fmt.Printf("failed to invalidate dashboard cache: %v\n", err)
	}

	categoriesKey := fmt.Sprintf("categories:%d", userID)
	if err := s.redis.Delete(ctx, categoriesKey); err != nil {
		fmt.Printf("failed to invalidate categories cache: %v\n", err)
	}
}
