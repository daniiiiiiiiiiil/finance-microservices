package ports

import (
	"time"

	"context"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/repository/postgres/pool"
)

type PoolInterface interface {
	Begin(ctx context.Context) (pool.Tx, error)
	OpTimeout() time.Duration
}
