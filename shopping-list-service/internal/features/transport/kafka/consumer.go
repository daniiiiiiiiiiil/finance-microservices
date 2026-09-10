package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/kafka"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/features/service/saga"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/pkg/logger"
	"go.uber.org/zap"
)

type ShoppingKafkaConsumer struct {
	consumer         *kafka.Consumer
	sagaOrchestrator *saga.DeleteUserSaga
	logger           *logger.Logger
}

func NewShoppingKafkaConsumer(
	consumer *kafka.Consumer,
	sagaOrchestrator *saga.DeleteUserSaga,
	logger *logger.Logger) *ShoppingKafkaConsumer {
	return &ShoppingKafkaConsumer{
		consumer:         consumer,
		sagaOrchestrator: sagaOrchestrator,
		logger:           logger,
	}
}

func (c *ShoppingKafkaConsumer) Start(ctx context.Context) error {
	c.consumer.RegisterHandler(kafka.EventTypeUserDeleted, c.handleUserDeleted)

	c.logger.Info("shopping kafka consumer started")

	return c.consumer.Start(ctx)
}

func (c *ShoppingKafkaConsumer) handleUserDeleted(ctx context.Context, event kafka.Event) error {
	var userDeletedEvent kafka.UserDeletedEvent
	if err := json.Unmarshal(event.Data, &userDeletedEvent); err != nil {
		c.logger.Error("failed to unmarshal user.deleted event",
			zap.String("event_id", event.ID),
			zap.Error(err))
		return fmt.Errorf("unmarshal user.deleted event: %w", err)
	}

	c.logger.Info("received user.deleted event",
		zap.Int("user_id", userDeletedEvent.UserID),
		zap.String("email", userDeletedEvent.Email),
		zap.String("event_id", event.ID))

	if err := c.sagaOrchestrator.HandleUserDeleted(ctx, userDeletedEvent.UserID); err != nil {
		c.logger.Error("failed to handle user.deleted event",
			zap.Int("user_id", userDeletedEvent.UserID),
			zap.String("event_id", event.ID),
			zap.Error(err))
		return fmt.Errorf("handle user.deleted: %w", err)
	}

	c.logger.Info("user.deleted event processed successfully",
		zap.Int("user_id", userDeletedEvent.UserID))

	return nil
}

func (c *ShoppingKafkaConsumer) Close() error {
	c.logger.Info("shopping kafka consumer closing")
	return c.consumer.Close()
}
