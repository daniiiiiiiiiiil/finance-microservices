package service

import (
	"fmt"
	"time"

	"context"

	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/kafka"
)

//func (s *FinanceService) sendTransactionEvent(ctx context.Context, eventType string, transaction domain.Finance) {
//	eventData := domain.TransactionEvent{
//		TransactionID:   transaction.ID,
//		UserID:          transaction.UserID,
//		TypeTransaction: transaction.TypeTransaction,
//		Amount:          transaction.Amount,
//		Category:        transaction.Category,
//		CreatedAt:       transaction.CreatedAt,
//	}
//	if err := s.eventPublisher.Publish(ctx, eventType, eventData); err != nil {
//		fmt.Printf("failed to send kafka event: %v\n", err)
//	}
//}

func (s *FinanceService) sendUserTransactionsDeletedEvent(ctx context.Context, userID int, count int) {
	eventData := domain.UserTransactionsDeletedEvent{
		UserID:       userID,
		DeletedCount: count,
		Timestamp:    time.Now(),
	}
	if err := s.eventPublisher.Publish(ctx, kafka.EventTypeUserTransactionsDeleted, eventData); err != nil {
		fmt.Printf("failed to send user transactions deleted event: %v\n", err)
	}
}
