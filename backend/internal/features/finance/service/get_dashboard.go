package service

import (
	"backend/internal/core/domain"
	"context"
	"fmt"
)

func (s *FinanceService) GetDashboard(ctx context.Context, userID int) (domain.Dashboard, error) {
	dashboard, err := s.repo.GetDashboard(ctx, userID)
	if err != nil {
		return domain.Dashboard{}, fmt.Errorf("get dashboard: %w", err)
	}
	return dashboard, nil
}
