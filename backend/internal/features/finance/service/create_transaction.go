package service

import (
	"backend/internal/core/domain"
	"backend/internal/core/kafka"
	"context"
	"fmt"
)

func (s *FinanceService) CreateTransaction(ctx context.Context, transaction domain.Finance) (domain.Finance, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Finance{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	if err := transaction.Validate(); err != nil {
		return domain.Finance{}, fmt.Errorf("validation failed: %w", err)
	}

	created, err := s.repo.CreateTransactionTx(ctx, tx, transaction)
	if err != nil {
		return domain.Finance{}, fmt.Errorf("create transaction: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Finance{}, fmt.Errorf("commit transaction: %w", err)
	}

	go s.invalidateCache(context.Background(), transaction.UserID)

	go s.sendTransactionEvent(context.Background(), kafka.EventTypeTransactionCreated, created)

	return created, nil
}
func (s *FinanceService) invalidateCache(ctx context.Context, userID int) {
	dashboardKey := fmt.Sprintf("dashboard:%d", userID)
	if err := s.redis.Delete(ctx, dashboardKey); err != nil {
		fmt.Printf("failed to invalidate dashboard cache: %v\n", err)
	}

	categoriesKey := fmt.Sprintf("categories:%d", userID)
	if err := s.redis.Delete(ctx, categoriesKey); err != nil {
		fmt.Printf("failed to invalidate categories cache: %v\n", err)
	}
}
