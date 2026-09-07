package service

import (
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/pkg/logger"
	"golang.org/x/net/context"
)

type ShoppingService struct {
	shoppingRepository ports.ShoppingListRepository
	pool               ports.PoolInterface
	shoppingCache      ports.ShoppingCacheInterface
	shoppingListCache  ports.ShoppingListCacheInterface
	redis              ports.RedisInterface
	logger             *logger.Logger
}

func NewShoppingService(
	shoppingRepository ports.ShoppingListRepository,
	pool ports.PoolInterface,
	shoppingCache ports.ShoppingCacheInterface,
	shoppingListCache ports.ShoppingListCacheInterface,
	redis ports.RedisInterface,
	logger *logger.Logger,
) *ShoppingService {
	return &ShoppingService{
		shoppingRepository: shoppingRepository,
		pool:               pool,
		shoppingCache:      shoppingCache,
		shoppingListCache:  shoppingListCache,
		redis:              redis,
		logger:             logger,
	}
}

func (s *ShoppingService) invalidateCache(ctx context.Context, userID int) {
	dashboardKey := fmt.Sprintf("dashboard:%d", userID)
	if err := s.redis.Delete(ctx, dashboardKey); err != nil {
		fmt.Printf("failed to invalidate dashboard cache: %v\n", err)
	}

	categoriesKey := fmt.Sprintf("categories:%d", userID)
	if err := s.redis.Delete(ctx, categoriesKey); err != nil {
		fmt.Printf("failed to invalidate categories cache: %v\n", err)
	}
}
