package service

import (
	"context"
	"fmt"
)

func (s *AdminService) GetMetrics(ctx context.Context) (Metrics, error) {
	metrics, err := s.repo.GetMetrics(ctx)
	if err != nil {
		return Metrics{}, fmt.Errorf("get metrics: %w", err)
	}
	return metrics, nil
}
