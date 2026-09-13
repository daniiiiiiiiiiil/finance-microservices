package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

func (s *ShoppingService) DeleteUserData(ctx context.Context, userID int) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != context.Canceled {
			s.logger.Error("rollback transaction", zap.Error(err))
		}
	}()

	shoppings, err := s.shoppingRepository.GetShoppingsByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user shoppings: %w", err)
	}

	var imageKeys []string
	var shoppingIDs []int
	for _, shopping := range shoppings {
		shoppingIDs = append(shoppingIDs, shopping.ID)
		if shopping.ImageKey != nil && *shopping.ImageKey != "" {
			imageKeys = append(imageKeys, *shopping.ImageKey)
		}
	}

	for _, imageKey := range imageKeys {
		if err := s.storage.Delete(ctx, imageKey); err != nil {
			s.logger.Warn("failed to delete image",
				zap.String("image_key", imageKey),
				zap.Error(err))
		}
	}

	for _, shoppingID := range shoppingIDs {
		if err := s.shoppingRepository.DeleteShopping(ctx, tx, shoppingID, userID); err != nil {
			return fmt.Errorf("delete shopping %d: %w", shoppingID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	s.logger.Info("user data deleted",
		zap.Int("user_id", userID),
		zap.Int("shopping_count", len(shoppingIDs)),
		zap.Int("images_count", len(imageKeys)))

	return nil
}
