package service

import (
	"context"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/pkg/pagination"
	"go.uber.org/zap"
)

func (s *ShoppingService) ListShopping(ctx context.Context, limit, offset int, userID int) ([]domain.Shopping, int, error) {
	limit, offset = pagination.LimitOffset(limit, offset)

	list, found := s.shoppingListCache.GetShoppingList(ctx, limit, offset)
	if found {
		total, err := s.shoppingRepository.GetTotalShopping(ctx, nil, userID)
		if err != nil {
			return nil, 0, fmt.Errorf("error getting total shopping list: %w", err)
		}
		return list, total, nil
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != context.Canceled {
			s.logger.Error("rollback transaction", zap.Error(err))
		}
	}()

	list, total, err := s.shoppingRepository.ListShopping(ctx, tx, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error listing shoppings: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, 0, fmt.Errorf("commit transaction: %w", err)
	}

	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.shoppingListCache.SetShoppingList(cacheCtx, list, limit, offset); err != nil {
			s.logger.Error("failed to cache shopping list",
				zap.Int("limit", limit),
				zap.Int("offset", offset),
				zap.Error(err))
		}
	}()

	return list, total, nil
}
