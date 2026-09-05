package service

import (
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/pkg/logger"
)

type ShoppingService struct {
	shoppingRepository ports.ShoppingListRepository
	pool               ports.PoolInterface
	logger             *logger.Logger
}

func NewShoppingService(
	shoppingRepository ports.ShoppingListRepository,
	pool ports.PoolInterface,
	logger *logger.Logger,
) *ShoppingService {
	return &ShoppingService{
		shoppingRepository: shoppingRepository,
		pool:               pool,
		logger:             logger,
	}
}
