package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func (s *FinanceService) GetCategories(ctx context.Context, userID int) ([]string, error) {
	key := fmt.Sprintf("categories:%d", userID)
	var categories []string
	err := s.redis.Get(ctx, key, &categories)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			categories, err = s.repo.GetCategories(ctx, userID)
			if err != nil {
				return nil, fmt.Errorf("failed to get categories: %w", err)
			}
			if err := s.redis.Set(ctx, key, categories, time.Hour*24); err != nil {
				fmt.Printf("failed to cache categories: %v\n", err)
			}
			return categories, nil
		}
		return nil, fmt.Errorf("Error getting categories for user %d", userID)
	}
	return categories, nil
}
