package service

import (
	"fmt"

	"context"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/domain"
	"go.uber.org/zap"
)

func (s *ShoppingService) CreateShopping(ctx context.Context, shopping domain.Shopping) (domain.Shopping, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.Shopping{}, fmt.Errorf("begin shopping transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			s.logger.Error("rollback shopping transaction", zap.Error(err))
		}
	}()

	if err := shopping.Validate(); err != nil {
		return domain.Shopping{}, fmt.Errorf("validate shopping data: %w", err)
	}

	created, err := s.shoppingRepository.CreateShopping(ctx, tx, shopping)
	if err != nil {
		return domain.Shopping{}, fmt.Errorf("create shopping data: %w", err)
	}

	return created, nil
}
