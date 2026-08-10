package service

import (
	"backend/internal/core/domain"
	"backend/internal/core/kafka"
	"context"
	"fmt"
)

func (s *FinanceService) sendTransactionEvent(ctx context.Context, eventType string, transaction domain.Finance) {
	if s.producer == nil {
		return
	}

	eventData := kafka.TransactionEvent{
		TransactionID:   transaction.ID,
		UserID:          transaction.UserID,
		TypeTransaction: transaction.TypeTransaction,
		Amount:          transaction.Amount,
		Category:        transaction.Category,
		CreatedAt:       transaction.CreatedAt,
	}
	if err := s.producer.SendEvent(ctx, eventType, eventData); err != nil {
		fmt.Printf("failed to send kafka event: %v\n", err)
	}
}
