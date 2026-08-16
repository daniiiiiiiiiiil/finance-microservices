package service

import (
	"backend/internal/core/domain"
	"backend/internal/core/kafka"
	"context"
	"fmt"
)

func (s *FinanceService) UpdateTransaction(ctx context.Context, transaction domain.Finance) (domain.Finance, error) {
	fmt.Printf("UpdateTransaction called with ID: %d\n", transaction.ID)
	existing, err := s.repo.GetTransaction(ctx, transaction.ID)
	if err != nil {
		return domain.Finance{}, fmt.Errorf("get existing transaction: %w", err)
	}

	fmt.Printf("UpdateTransaction called with ID: %d\n", transaction.ID)

	transaction.Version = existing.Version
	transaction.UserID = existing.UserID
	transaction.CreatedAt = existing.CreatedAt

	if err := transaction.Validate(); err != nil {
		return domain.Finance{}, fmt.Errorf("validation failed: %w", err)
	}

	updated, err := s.repo.UpdateTransaction(ctx, transaction)
	if err != nil {
		return domain.Finance{}, fmt.Errorf("update transaction: %w", err)
	}

	go s.invalidateCache(context.Background(), transaction.UserID)

	go s.sendTransactionEvent(context.Background(), kafka.EventTypeTransactionUpdated, updated)

	return updated, nil
}
