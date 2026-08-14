package service

import (
	service "backend/internal/features/admin/service"
	"context"
	"fmt"
)

func (s *FinanceService) GetMetrics(ctx context.Context) (service.Metrics, error) {
	metrics, err := s.repo.GetMetrics(ctx)
	if err != nil {
		return service.Metrics{}, fmt.Errorf("get metrics: %w", err)
	}
	return metrics, nil
}
