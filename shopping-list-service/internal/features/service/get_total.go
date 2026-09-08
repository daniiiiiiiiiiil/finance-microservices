package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

func (s *ShoppingService) GetTotal(ctx context.Context, userID int) (int, error) {
	total, err := s.shoppingRepository.GetTotalShopping(ctx, nil, userID)
	if err != nil {
		s.logger.Error("Get total shopping error", zap.Error(err))
		return 0, fmt.Errorf("shopping service get total: %w", err)
	}
	return total, nil
}
