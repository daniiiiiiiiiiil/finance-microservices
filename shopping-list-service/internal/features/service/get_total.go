package service

import (
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/net/context"
)

func (s *ShoppingService) GetTotal(ctx context.Context, userID int) (int, error) {
	total, err := s.shoppingRepository.GetTotalShopping(ctx, userID)
	if err != nil {
		s.logger.Error("Get total shopping error", zap.Error(err))
		return 0, fmt.Errorf("shopping service get total: %w", err)
	}
	return total, nil
}
