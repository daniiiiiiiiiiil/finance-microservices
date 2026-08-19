package service

import (
	"context"
	"fmt"
	"time"
)

func (s *FinanceService) GetTransactionsCount(
	ctx context.Context,
	userID int,
	transactionType, category *string,
	from, to *time.Time,
) (int, error) {
	total, err := s.repo.GetTransactionsCount(ctx, userID, transactionType, category, from, to)
	if err != nil {
		return 0, fmt.Errorf("get transactions count: %w", err)
	}
	return total, nil
}