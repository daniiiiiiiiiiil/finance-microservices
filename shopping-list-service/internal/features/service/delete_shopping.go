package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

func (s *ShoppingService) DeleteShoppingList(ctx context.Context, id int, userID int) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != context.Canceled {
			s.logger.Error("rollback transaction", zap.Error(err))
		}
	}()

	shopping, err := s.shoppingRepository.GetShopping(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("GetShoppingList: %w", err)
	}
	if id <= 0 {
		return fmt.Errorf("shopping list id must be positive")
	}

	if shopping.ImageKey != nil && *shopping.ImageKey != "" {
		if err := s.storage.Delete(ctx, *shopping.ImageKey); err != nil {
			s.logger.Warn("failed to delete image from S3",
				zap.Int("shopping_id", id),
				zap.String("image_key", *shopping.ImageKey),
				zap.Error(err))
		} else {
			s.logger.Debug("image deleted from S3",
				zap.Int("shopping_id", id),
				zap.String("image_key", *shopping.ImageKey))
		}
	}

	if err := s.shoppingRepository.DeleteShopping(ctx, tx, id, userID); err != nil {
		s.logger.Error("shopping service delete shopping list failed", zap.Error(err))
		return fmt.Errorf("shopping service delete: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	go s.invalidateCache(ctx, userID)
	return nil
}
