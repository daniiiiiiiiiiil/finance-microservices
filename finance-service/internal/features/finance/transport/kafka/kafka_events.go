package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/kafka"
)

func (p *FinanceEventPublisher) SendTransactionEvent(ctx context.Context, eventType string, transaction domain.Finance) {
	eventData := kafka.TransactionEvent{
		TransactionID:   transaction.ID,
		UserID:          transaction.UserID,
		TypeTransaction: transaction.TypeTransaction,
		Amount:          transaction.Amount,
		Category:        transaction.Category,
		CreatedAt:       transaction.CreatedAt,
	}
	if err := p.Publish(ctx, eventType, eventData); err != nil {
		fmt.Printf("failed to send kafka event: %v\n", err)
	}
}

func (p *FinanceEventPublisher) SendUserTransactionsDeletedEvent(ctx context.Context, userID int, count int) {
	eventData := kafka.UserTransactionsDeletedEvent{
		UserID:       userID,
		DeletedCount: count,
		Timestamp:    time.Now(),
	}
	if err := p.Publish(ctx, kafka.EventTypeUserTransactionsDeleted, eventData); err != nil {
		fmt.Printf("failed to send user transactions deleted event: %v\n", err)
	}
}
