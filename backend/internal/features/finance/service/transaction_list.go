package service

import (
	"backend/internal/core/domain"
	"backend/pkg/pagination"
	"fmt"
	"time"

	"golang.org/x/net/context"
)

func (s *FinanceService) GetTransactions(
	ctx context.Context,
	userID int,
	transactionType, category *string,
	from, to *time.Time,
	limit, offset int,
) ([]domain.Finance, error) {
	limit, offset = pagination.LimitOffset(limit, offset)

	if from != nil && to != nil && from.After(*to) {
		return nil, fmt.Errorf("from date must be before to date")
	}

	transactions, err := s.repo.GetTransactions(ctx, userID, transactionType, category, from, to, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get transactions: %w", err)
	}

	return transactions, nil
}

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
