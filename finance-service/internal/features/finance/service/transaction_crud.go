package service

import (
	"fmt"

	"context"

	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/kafka"
	"go.uber.org/zap"
)

func (s *FinanceService) CreateTransaction(ctx context.Context, transaction domain.Finance) (domain.Finance, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Finance{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != context.Canceled {
			s.logger.Error("failed to rollback transaction", zap.Error(err))
		}
	}()
	if err := transaction.Validate(); err != nil {
		return domain.Finance{}, fmt.Errorf("validation failed: %w", err)
	}

	created, err := s.repo.CreateTransactionTx(ctx, tx, transaction)
	if err != nil {
		return domain.Finance{}, fmt.Errorf("create transaction: %w", err)
	}

	event, err := domain.NewOutboxEvent(
		int64(created.ID),
		"transaction",
		kafka.EventTypeTransactionCreated,
		created,
	)
	if err != nil {
		return domain.Finance{}, fmt.Errorf("create outbox event: %w", err)
	}

	if err := s.outboxRepo.SaveTx(ctx, tx, event); err != nil {
		return domain.Finance{}, fmt.Errorf("save outbox event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Finance{}, fmt.Errorf("commit transaction: %w", err)
	}

	go s.invalidateCache(context.Background(), transaction.UserID)

	//go s.sendTransactionEvent(context.Background(), kafka.EventTypeTransactionCreated, created)

	return created, nil
}

func (s *FinanceService) GetTransaction(ctx context.Context, id int) (domain.Finance, error) {
	transaction, err := s.repo.GetTransaction(ctx, id)
	if err != nil {
		return domain.Finance{}, fmt.Errorf("get transaction: %w", err)
	}
	return transaction, nil
}

func (s *FinanceService) UpdateTransaction(ctx context.Context, transaction domain.Finance) (domain.Finance, error) {
	fmt.Printf("UpdateTransaction called with ID: %d\n", transaction.ID)
	existing, err := s.repo.GetTransaction(ctx, transaction.ID)
	if err != nil {
		return domain.Finance{}, fmt.Errorf("get existing transaction: %w", err)
	}

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

	event, err := domain.NewOutboxEvent(
		int64(updated.ID),
		"transaction",
		kafka.EventTypeTransactionUpdated,
		updated,
	)
	if err != nil {
		return domain.Finance{}, fmt.Errorf("create outbox event: %w", err)
	}

	if err := s.outboxRepo.Save(ctx, event); err != nil {
		return domain.Finance{}, fmt.Errorf("save outbox event: %w", err)
	}

	go s.invalidateCache(context.Background(), transaction.UserID)

	//go s.sendTransactionEvent(context.Background(), kafka.EventTypeTransactionUpdated, updated)

	return updated, nil
}

func (s *FinanceService) DeleteTransaction(ctx context.Context, id int) error {
	tx, err := s.repo.GetTransaction(ctx, id)
	if err != nil {
		return fmt.Errorf("get transaction: %w", err)
	}

	if err := s.repo.DeleteTransaction(ctx, id); err != nil {
		return fmt.Errorf("delete transaction: %w", err)
	}

	event, err := domain.NewOutboxEvent(
		int64(tx.ID),
		"transaction",
		kafka.EventTypeTransactionDeleted,
		tx,
	)
	if err != nil {
		return fmt.Errorf("create outbox event: %w", err)
	}

	if err := s.outboxRepo.Save(ctx, event); err != nil {
		return fmt.Errorf("save outbox event: %w", err)
	}

	go s.invalidateCache(context.Background(), tx.UserID)

	//go s.sendTransactionEvent(context.Background(), kafka.EventTypeTransactionDeleted, tx)

	return nil
}
