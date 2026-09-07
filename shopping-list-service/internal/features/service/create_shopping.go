package service

import (
	"context"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/domain"
	"go.uber.org/zap"
)

func (s *ShoppingService) CreateShopping(ctx context.Context, shopping domain.Shopping, userID int) (domain.Shopping, error) {
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

	if err := tx.Commit(ctx); err != nil {
		return domain.Shopping{}, fmt.Errorf("commit transaction: %w", err)
	}

	return created, nil
}
