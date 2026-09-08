package service

import (
	"context"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/domain"
	"go.uber.org/zap"
)

func (s *ShoppingService) GetShopping(ctx context.Context, id int, userID int) (domain.Shopping, error) {
	shopping, err := s.shoppingCache.GetShopping(ctx, id)
	if err == nil {
		return shopping, nil
	}

	shopping, err = s.shoppingRepository.GetShopping(ctx, id, userID)
	if err != nil {
		return domain.Shopping{}, fmt.Errorf("error getting shopping from repository: %w", err)
	}

	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.shoppingCache.SetShopping(cacheCtx, shopping); err != nil {
			s.logger.Error("failed to cache shopping",
				zap.Int("id", id),
				zap.Error(err))
		}
	}()

	return shopping, nil
}
