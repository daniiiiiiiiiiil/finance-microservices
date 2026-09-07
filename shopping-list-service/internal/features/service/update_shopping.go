package service

import (
	"context"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/domain"
	"go.uber.org/zap"
)

func (s *ShoppingService) UpdateShopping(ctx context.Context, shopping *domain.Shopping, userID int) (domain.Shopping, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Shopping{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			s.logger.Error("rollback transaction", zap.Error(err))
		}
	}()

	exists, err := s.shoppingRepository.GetShopping(ctx, shopping.ID, userID)
	if err != nil {
		return domain.Shopping{}, fmt.Errorf("error getting shopping with id %d: %w", shopping.ID, err)
	}
	shopping.Version = exists.Version
	shopping.CreatedAt = exists.CreatedAt

	if err := shopping.Validate(); err != nil {
		return domain.Shopping{}, fmt.Errorf("error validating shopping with id %d: %w", shopping.ID, err)
	}

	updated, err := s.shoppingRepository.UpdateShopping(ctx, shopping, userID)
	if err != nil {
		return domain.Shopping{}, fmt.Errorf("error updating shopping with id %d: %w", shopping.ID, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Shopping{}, fmt.Errorf("commit transaction: %w", err)
	}

	go s.invalidateCache(ctx, shopping.ID)

	return updated, nil
}
