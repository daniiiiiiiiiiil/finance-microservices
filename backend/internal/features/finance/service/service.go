package service

import (
	"backend/internal/core/cache"
	"backend/internal/core/domain"
	"backend/internal/core/kafka"
	"backend/internal/core/repository/postgres/pool"
	"backend/internal/features/admin/service"
	"backend/pkg/logger"
	"context"
	"time"
)

var _ FinanceServiceInterface = (*FinanceService)(nil)

//go:generate mockgen -destination=mocks/mock_finance_service.go -package=mocks -source=service.go FinanceService
type FinanceRepository interface {
	CreateTransactionTx(ctx context.Context, tx pool.Tx, transaction domain.Finance) (domain.Finance, error)
	GetTransaction(ctx context.Context, id int) (domain.Finance, error)
	GetTransactions(ctx context.Context, userID int, transactionType, category *string, from, to *time.Time, limit, offset int) ([]domain.Finance, error)
	UpdateTransaction(ctx context.Context, transaction domain.Finance) (domain.Finance, error)
	DeleteTransaction(ctx context.Context, id int) error
	GetCategories(ctx context.Context, userID int) ([]string, error)
	GetDashboard(ctx context.Context, userID int) (domain.Dashboard, error)
	DeleteUserTransactions(ctx context.Context, userID int) (int, error)
	GetMetrics(ctx context.Context) (service_admin.Metrics, error)
	GetTransactionsCount(ctx context.Context, userID int, transactionType, category *string, from, to *time.Time) (int, error)
}

type FinanceService struct {
	repo     FinanceRepository
	pool     pool.Pool
	redis    cache.RedisInterface
	producer *kafka.Producer
	logger   *logger.Logger
}

func NewFinanceService(repo FinanceRepository, pool pool.Pool, redis cache.RedisInterface, producer *kafka.Producer, logger *logger.Logger) *FinanceService {
	return &FinanceService{repo: repo, pool: pool, redis: redis, producer: producer, logger: logger}
}
