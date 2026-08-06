package service

import (
	"backend/internal/core/domain"
	"context"
	"fmt"
)

func (s *FinanceService) GetTransaction(ctx context.Context, id int) (domain.Finance, error) {
	transaction, err := s.repo.GetTransaction(ctx, id)
	if err != nil {
		return domain.Finance{}, fmt.Errorf("get transaction: %w", err)
	}
	return transaction, nil
}
