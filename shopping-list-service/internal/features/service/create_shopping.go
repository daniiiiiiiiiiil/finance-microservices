package service

import (
	"context"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/domain"
	"go.uber.org/zap"
)

func (s *ShoppingService) CreateShopping(ctx context.Context, shopping domain.Shopping, userID int, fileData []byte, filename string) (domain.Shopping, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Shopping{}, fmt.Errorf("begin shopping transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != context.Canceled {
			s.logger.Error("rollback shopping transaction", zap.Error(err))
		}
	}()

	if err := shopping.Validate(); err != nil {
		return domain.Shopping{}, fmt.Errorf("validate shopping data: %w", err)
	}

	created, err := s.shoppingRepository.CreateShopping(ctx, tx, shopping, userID)
	if err != nil {
		return domain.Shopping{}, fmt.Errorf("create shopping data: %w", err)
	}

	if fileData != nil && len(fileData) > 0 && filename != "" {
		if err := validateImage(fileData, filename); err != nil {
			return domain.Shopping{}, fmt.Errorf("validate image: %w", err)
		}
		imageKey := generateImageKey(userID, created.ID, filename)
		if err := s.storage.Put(ctx, imageKey, fileData); err != nil {
			s.logger.Error("failed to upload image to S3",
				zap.Int("shopping_id", created.ID),
				zap.Int("user_id", userID),
				zap.String("key", imageKey),
				zap.Error(err))
			return domain.Shopping{}, fmt.Errorf("put image: %w", err)
		}
		created.ImageKey = &imageKey

		updated, err := s.shoppingRepository.UpdateShopping(ctx, tx, &created, userID)
		if err != nil {
			if delErr := s.storage.Delete(ctx, imageKey); delErr != nil {
				s.logger.Error("failed to delete image from S3",
					zap.Int("shopping_id", created.ID),
					zap.Int("user_id", userID),
					zap.String("key", imageKey),
					zap.Error(delErr))
			}
			return domain.Shopping{}, fmt.Errorf("update shopping data: %w", err)
		}
		created = updated

		s.logger.Info("image uploaded successfully",
			zap.Int("shopping_id", created.ID),
			zap.String("image_key", imageKey),
			zap.Int("size", len(fileData)))
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Shopping{}, fmt.Errorf("commit transaction: %w", err)
	}

	return created, nil
}
