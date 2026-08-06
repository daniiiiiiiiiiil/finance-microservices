package service

import (
	"backend/internal/core/domain"
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
	return created, nil
}
