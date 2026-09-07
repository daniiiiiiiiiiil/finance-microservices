package service

import (
	"context"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/pkg/pagination"
)

func (s *ShoppingService) ListShopping(ctx context.Context, limit, offset int, userID int) ([]domain.Shopping, int, error) {
	limit, offset = pagination.LimitOffset(limit, offset)

	list, found := s.shoppingListCache.GetShoppingList(ctx, limit, offset)
	if found {
		total, err := s.shoppingRepository.GetTotalShopping(ctx, userID)
		if err != nil {
			return nil, 0, fmt.Errorf("error getting total shopping list: %w", err)
		}
		return list, total, nil
	}

	list, total, err := s.shoppingRepository.ListShopping(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error listing shoppings: %w", err)
	}

	go func() {
		if err := s.shoppingListCache.SetShoppingList(context.Background(), list, limit, offset); err != nil {
			s.logger.Error("failed to cache shopping list: " + err.Error())
		}
	}()

	return list, total, nil
}
