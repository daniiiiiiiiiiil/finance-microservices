package service

import (
	"context"
	"fmt"

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
		if err := s.shoppingCache.SetShopping(context.Background(), shopping); err != nil {
			s.logger.Error("failed to cache shopping", zap.Int("id", id), zap.Error(err))
		}
	}()

	return shopping, nil
}
