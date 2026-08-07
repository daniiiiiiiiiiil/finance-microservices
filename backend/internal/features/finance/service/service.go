package service

import (
	"backend/internal/core/cache"
	"backend/internal/core/domain"
	"backend/internal/core/repository/postgres/pool"
	"context"
	"time"
)

//go:generate mockgen -destination=mocks/mock_finance_service.go -package=mocks -source=service.go FinanceService
type FinanceRepository interface {
	CreateTransactionTx(ctx context.Context, tx pool.Tx, transaction domain.Finance) (domain.Finance, error)
	GetTransaction(ctx context.Context, id int) (domain.Finance, error)
	GetTransactions(ctx context.Context, userID int, transactionType, category *string, from, to *time.Time, limit, offset int) ([]domain.Finance, error)
	UpdateTransaction(ctx context.Context, transaction domain.Finance) (domain.Finance, error)
	DeleteTransaction(ctx context.Context, id int) error
	GetCategories(ctx context.Context, userID int) ([]string, error)
	GetDashboard(ctx context.Context, userID int) (domain.Dashboard, error)
}

type FinanceService struct {
	repo  FinanceRepository
	pool  pool.Pool
	redis cache.RedisInterface
}

func NewFinanceService(repo FinanceRepository, pool pool.Pool, redis cache.RedisInterface) *FinanceService {
	return &FinanceService{repo: repo, pool: pool, redis: redis}
}
