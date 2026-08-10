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
