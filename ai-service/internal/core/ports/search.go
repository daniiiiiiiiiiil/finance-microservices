package ports

import (
	"context"

	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/core/domain"
)

type SearchClient interface {
	SearchStores(ctx context.Context, item domain.Item, nearAddress string) ([]domain.Store, error)
}
