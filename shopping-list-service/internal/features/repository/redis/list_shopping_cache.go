package redis

import (
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/cache"
)

type ShoppingListCache struct {
	client *cache.RedisClient
}

func NewShoppingListCache(client *cache.RedisClient) *ShoppingListCache {
	return &ShoppingListCache{
		client: client,
	}
}
