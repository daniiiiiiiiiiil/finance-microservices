package service

import (
	"backend/internal/core/kafka"
	"context"
	"fmt"
)

func (s *FinanceService) DeleteTransaction(ctx context.Context, id int) error {
	tx, err := s.repo.GetTransaction(ctx, id)
	if err != nil {
		return fmt.Errorf("get transaction: %w", err)
	}

	if err := s.repo.DeleteTransaction(ctx, id); err != nil {
		return fmt.Errorf("delete transaction: %w", err)
	}

	go s.invalidateCache(context.Background(), tx.UserID)

	go s.sendTransactionEvent(context.Background(), kafka.EventTypeTransactionDeleted, tx)

	return nil
}
