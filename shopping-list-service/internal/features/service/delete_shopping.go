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
		if err := tx.Rollback(ctx); err != nil {
			s.logger.Error("rollback transaction", zap.Error(err))
		}
	}()

	_, err = s.shoppingRepository.GetShopping(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("GetShoppingList: %w", err)
	}
	if id <= 0 {
		return fmt.Errorf("shopping list id must be positive")
	}

	if err := s.shoppingRepository.DeleteShopping(ctx, id, userID); err != nil {
		s.logger.Error("shopping service delete shopping list failed", zap.Error(err))
		return fmt.Errorf("shopping service delete: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
