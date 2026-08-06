package service

import (
	"context"
	"fmt"
)

func (s *FinanceService) GetCategories(ctx context.Context, userID int) ([]string, error) {
	category, err := s.repo.GetCategories(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("GetCategories: %w", err)
	}
	return category, nil
}
