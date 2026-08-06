package service

import (
	"backend/internal/core/domain"
	"context"
	"fmt"
)

func (s *FinanceService) UpdateTransaction(ctx context.Context, transaction domain.Finance) (domain.Finance, error) {
	existing, err := s.repo.GetTransaction(ctx, transaction.ID)
	if err != nil {
		return domain.Finance{}, fmt.Errorf("get existing transaction: %w", err)
	}

	transaction.Version = existing.Version

	if err := transaction.Validate(); err != nil {
		return domain.Finance{}, fmt.Errorf("validation failed: %w", err)
	}

	updated, err := s.repo.UpdateTransaction(ctx, transaction)
	if err != nil {
		return domain.Finance{}, fmt.Errorf("update transaction: %w", err)
	}

	return updated, nil
}
