package service

import (
	"context"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/domain"
	"go.uber.org/zap"
)

func (s *ShoppingService) UpdateShopping(ctx context.Context, shopping *domain.Shopping, userID int, fileData []byte, filename string) (domain.Shopping, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Shopping{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != context.Canceled {
			s.logger.Error("rollback transaction", zap.Error(err))
		}
	}()

	exists, err := s.shoppingRepository.GetShopping(ctx, shopping.ID, userID)
	if err != nil {
		return domain.Shopping{}, fmt.Errorf("error getting shopping with id %d: %w", shopping.ID, err)
	}
	oldImageKey := exists.ImageKey
	shopping.Version = exists.Version
	shopping.CreatedAt = exists.CreatedAt

	if err := shopping.Validate(); err != nil {
		return domain.Shopping{}, fmt.Errorf("error validating shopping with id %d: %w", shopping.ID, err)
	}

	updated, err := s.shoppingRepository.UpdateShopping(ctx, tx, shopping, userID)
	if err != nil {
		return domain.Shopping{}, fmt.Errorf("error updating shopping with id %d: %w", shopping.ID, err)
	}

	if len(fileData) > 0 && filename != "" {
		if err := validateImage(fileData, filename); err != nil {
			return domain.Shopping{}, fmt.Errorf("validate image: %w", err)
		}
		imageKey := generateImageKey(userID, updated.ID, filename)
		if err := s.storage.Put(ctx, imageKey, fileData); err != nil {
			s.logger.Error("put image to storage",
				zap.Int("shopping_id", updated.ID),
				zap.Int("user_id", userID),
				zap.String("key", imageKey),
				zap.Error(err))
			return domain.Shopping{}, fmt.Errorf("error putting file with id %d: %w", updated.ID, err)
		}
		updated.ImageKey = &imageKey

		updatedImg, err := s.shoppingRepository.UpdateShopping(ctx, tx, &updated, userID)
		if err != nil {
			if delErr := s.storage.Delete(ctx, imageKey); delErr != nil {
				s.logger.Error("delete image from storage",
					zap.Int("shopping_id", updated.ID),
					zap.Int("user_id", userID),
					zap.String("key", imageKey),
					zap.Error(delErr))
			}
			s.logger.Error("put image to storage",
				zap.Int("shopping_id", updated.ID),
				zap.Int("user_id", userID),
				zap.String("key", imageKey),
				zap.Error(err))
			return domain.Shopping{}, fmt.Errorf("error putting file with id %d: %w", updated.ID, err)
		}
		updated = updatedImg

		if oldImageKey != updated.ImageKey && *oldImageKey != "" {
			if err := s.storage.Delete(ctx, *oldImageKey); err != nil {
				s.logger.Error("delete old image from storage",
					zap.Int("shopping_id", updated.ID),
					zap.Int("user_id", userID),
					zap.String("key", imageKey),
					zap.Error(err))
			} else {
				s.logger.Info("delete old image from storage", zap.Int("shopping_id", updated.ID), zap.String("old_key", *oldImageKey))
			}
		}

		s.logger.Info("updated shopping with id %d",
			zap.Int("shopping_id", updated.ID),
			zap.Int("user_id", userID),
			zap.Int("size", len(fileData)))
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Shopping{}, fmt.Errorf("commit transaction: %w", err)
	}

	go s.invalidateCache(ctx, userID)

	return updated, nil
}
