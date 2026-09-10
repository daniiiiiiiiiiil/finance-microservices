package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/kafka"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/pkg/logger"
	"go.uber.org/zap"
)

type ShoppingEventPublisher struct {
	producer *kafka.Producer
	logger   *logger.Logger
}

func NewShoppingEventPublisher(producer *kafka.Producer, logger *logger.Logger) *ShoppingEventPublisher {
	return &ShoppingEventPublisher{
		producer: producer,
		logger:   logger,
	}
}

func (p *ShoppingEventPublisher) Publish(ctx context.Context, eventType string, data interface{}) error {
	if err := p.producer.SendEvent(ctx, eventType, data); err != nil {
		p.logger.Error("error publishing event", zap.Error(err))
		return fmt.Errorf("error publishing event: %w", err)
	}
	return nil
}

func (p *ShoppingEventPublisher) SendShoppingDeletedEvent(ctx context.Context, userID int, deletedCount int, deletedIDs []int, imageKeys []string) error {
	event := kafka.ShoppingDeletedEvent{
		UserID:       userID,
		DeletedCount: deletedCount,
		DeletedIDs:   deletedIDs,
		ImageKeys:    imageKeys,
		DeletedAt:    time.Now(),
	}
	p.logger.Debug("sending shopping.deleted event",
		zap.Int("user_id", userID),
		zap.Int("deleted_count", deletedCount))

	return p.Publish(ctx, kafka.EventTypeShoppingDeleted, event)
}

func (p *ShoppingEventPublisher) SendDeleteCompletedEvent(ctx context.Context, userID int, deletedShoppingCount int, deletedImagesCount int) error {
	event := kafka.UserDeleteCompletedEvent{
		UserID:               userID,
		Status:               "completed",
		DeletedShoppingCount: deletedShoppingCount,
		DeletedImagesCount:   deletedImagesCount,
		CompletedAt:          time.Now()}
	p.logger.Info("sending user.delete.completed event",
		zap.Int("user_id", userID),
		zap.Int("deleted_shopping_count", deletedShoppingCount),
		zap.Int("deleted_images_count", deletedImagesCount))

	return p.Publish(ctx, kafka.EventTypeUserDeleteCompleted, event)
}

func (p *ShoppingEventPublisher) SendDeleteFailedEvent(ctx context.Context, userID int, reason string, err error, failedStep string) error {
	event := kafka.UserDeleteFailedEvent{
		UserID:     userID,
		Reason:     reason,
		Error:      err.Error(),
		FailedStep: failedStep,
		RetryCount: 0,
		FailedAt:   time.Now(),
	}

	p.logger.Error("sending user.delete.failed event",
		zap.Int("user_id", userID),
		zap.String("reason", reason),
		zap.Error(err))

	return p.Publish(ctx, kafka.EventTypeUserDeleteFailed, event)
}

func (p *ShoppingEventPublisher) SendShoppingCreatedEvent(ctx context.Context, shoppingID int, userID int, title string, amountNow, amountFinish float64, imageKey *string) error {
	event := kafka.ShoppingCreatedEvent{
		ShoppingID:   shoppingID,
		UserID:       userID,
		Title:        title,
		AmountNow:    amountNow,
		AmountFinish: amountFinish,
		ImageKey:     imageKey,
		CreatedAt:    time.Now(),
	}

	p.logger.Debug("sending shopping.created event",
		zap.Int("shopping_id", shoppingID),
		zap.Int("user_id", userID))

	return p.Publish(ctx, kafka.EventTypeShoppingCreated, event)
}

func (p *ShoppingEventPublisher) SendShoppingUpdatedEvent(ctx context.Context, shoppingID int, userID int, oldValues, newValues map[string]interface{}) error {
	event := kafka.ShoppingUpdatedEvent{
		ShoppingID: shoppingID,
		UserID:     userID,
		OldValues:  oldValues,
		NewValues:  newValues,
		UpdatedAt:  time.Now(),
	}

	p.logger.Debug("sending shopping.updated event",
		zap.Int("shopping_id", shoppingID),
		zap.Int("user_id", userID))

	return p.Publish(ctx, kafka.EventTypeShoppingUpdated, event)
}

func (p *ShoppingEventPublisher) SendImageUploadedEvent(ctx context.Context, shoppingID int, userID int, imageKey string, size int64, format string) error {
	event := kafka.ImageUploadedEvent{
		ShoppingID: shoppingID,
		UserID:     userID,
		ImageKey:   imageKey,
		Size:       size,
		Format:     format,
		UploadedAt: time.Now(),
	}

	p.logger.Debug("sending shopping.image.uploaded event",
		zap.Int("shopping_id", shoppingID),
		zap.Int("user_id", userID),
		zap.String("image_key", imageKey))

	return p.Publish(ctx, kafka.EventTypeImageUploaded, event)
}

func (p *ShoppingEventPublisher) SendImageDeletedEvent(ctx context.Context, shoppingID int, userID int, imageKey string) error {
	event := kafka.ImageDeletedEvent{
		ShoppingID: shoppingID,
		UserID:     userID,
		ImageKey:   imageKey,
		DeletedAt:  time.Now(),
	}

	p.logger.Debug("sending shopping.image.deleted event",
		zap.Int("shopping_id", shoppingID),
		zap.Int("user_id", userID),
		zap.String("image_key", imageKey))

	return p.Publish(ctx, kafka.EventTypeImageDeleted, event)
}
