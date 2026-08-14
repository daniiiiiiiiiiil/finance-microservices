package service

import (
	"context"
	"fmt"
)

func (s *FinanceService) DeleteUserTransactions(ctx context.Context, userID int) (int, error) {
	count, err := s.repo.DeleteUserTransactions(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("error deleting user transactions: %w", err)
	}

	go s.invalidateCache(context.Background(), userID)

	go s.sendUserTransactionsDeletedEvent(context.Background(), userID, count)

	return count, nil
}
