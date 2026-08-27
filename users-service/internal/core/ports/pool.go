package ports

import (
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/repository/postgres/pool"
	"golang.org/x/net/context"
)

type PoolInterface interface {
	Begin(ctx context.Context) (pool.Tx, error)
	OpTimeout() time.Duration
}
