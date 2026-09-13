package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/kafka"
	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/pkg/logger"
	"go.uber.org/zap"
)

type SagaEventPublisher struct {
	producer *kafka.Producer
	logger   *logger.Logger
}

func NewSagaEventPublisher(
	producer *kafka.Producer,
	logger *logger.Logger,
) *SagaEventPublisher {
	return &SagaEventPublisher{
		producer: producer,
		logger:   logger,
	}
}

func (p *SagaEventPublisher) Publish(ctx context.Context, eventType string, data interface{}) error {
	if err := p.producer.SendEvent(ctx, eventType, data); err != nil {
		p.logger.Error("failed to publish event",
			zap.String("event_type", eventType),
			zap.Error(err))
		return fmt.Errorf("publish event: %w", err)
	}
	return nil
}

func (p *SagaEventPublisher) SendUserMarkDeleting(ctx context.Context, userID int, sagaID string) error {
	event := UserMarkDeletingEvent{
		UserID:    userID,
		SagaID:    sagaID,
		Timestamp: time.Now(),
	}
	p.logger.Info("sending user.mark_deleting event",
		zap.Int("user_id", userID),
		zap.String("saga_id", sagaID))

	return p.Publish(ctx, EventTypeUserMarkDeleting, event)
}

func (p *SagaEventPublisher) SendUserRestore(ctx context.Context, userID int, sagaID string) error {
	event := UserRestoreEvent{
		UserID:    userID,
		SagaID:    sagaID,
		Timestamp: time.Now(),
	}

	p.logger.Info("sending user.restore event",
		zap.Int("user_id", userID),
		zap.String("saga_id", sagaID))
	return p.Publish(ctx, EventTypeUserRestore, event)
}

func (p *SagaEventPublisher) SendUserFinalizeDelete(ctx context.Context, userID int, sagaID string) error {
	event := UserFinalizeDeleteEvent{
		UserID:    userID,
		SagaID:    sagaID,
		Timestamp: time.Now(),
	}
	p.logger.Info("sending user.finalize delete event",
		zap.Int("user_id", userID),
		zap.String("saga_id", sagaID))
	return p.Publish(ctx, EventTypeUserFinalizeDelete, event)
}

func (p *SagaEventPublisher) SendUserDeleteCompleted(ctx context.Context, userID int, sagaID string) error {
	event := UserDeleteCompletedEvent{
		UserID:      userID,
		SagaID:      sagaID,
		CompletedAt: time.Now(),
	}

	p.logger.Info("sending user.delete.completed event",
		zap.Int("user_id", userID),
		zap.String("saga_id", sagaID))

	return p.Publish(ctx, EventTypeUserDeleteCompleted, event)
}

func (p *SagaEventPublisher) SendUserDeleteFailed(ctx context.Context, userID int, sagaID string, reason string, errMsg string) error {
	event := UserDeleteFailedEvent{
		UserID:   userID,
		SagaID:   sagaID,
		Reason:   reason,
		Error:    errMsg,
		FailedAt: time.Now(),
	}

	p.logger.Error("sending user.delete.failed event",
		zap.Int("user_id", userID),
		zap.String("saga_id", sagaID),
		zap.String("reason", reason))

	return p.Publish(ctx, EventTypeUserDeleteFailed, event)
}

func (p *SagaEventPublisher) SendDeleteShoppingData(ctx context.Context, userID int, sagaID string) error {
	event := ShoppingDeleteUserDataEvent{
		UserID:    userID,
		SagaID:    sagaID,
		Timestamp: time.Now(),
	}
	p.logger.Info("sending shopping_delete_user_data event",
		zap.Int("user_id", userID),
		zap.String("saga_id", sagaID))
	return p.Publish(ctx, EventTypeShoppingDeleteUserData, event)
}

func (p *SagaEventPublisher) SendDeleteFinanceTransactions(ctx context.Context, userID int, sagaID string) error {
	event := FinanceDeleteUserTransactionsEvent{
		UserID:    userID,
		SagaID:    sagaID,
		Timestamp: time.Now(),
	}

	p.logger.Info("sending finance.delete_user_transactions event",
		zap.Int("user_id", userID),
		zap.String("saga_id", sagaID))

	return p.Publish(ctx, EventTypeFinanceDeleteUserTransactions, event)
}
