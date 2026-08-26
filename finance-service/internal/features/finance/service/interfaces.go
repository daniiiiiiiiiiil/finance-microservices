package service

import (
	"context"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/features/finance/repository/postgres"
)

type FinanceServiceInterface interface {
	CreateTransaction(ctx context.Context, transaction domain.Finance) (domain.Finance, error)
	GetTransaction(ctx context.Context, id int) (domain.Finance, error)
	GetTransactions(ctx context.Context, userID int, transactionType, category *string, from, to *time.Time, limit, offset int) ([]domain.Finance, error)
	GetTransactionsCount(ctx context.Context, userID int, transactionType, category *string, from, to *time.Time) (int, error)
	UpdateTransaction(ctx context.Context, transaction domain.Finance) (domain.Finance, error)
	DeleteTransaction(ctx context.Context, id int) error
	GetCategories(ctx context.Context, userID int) ([]string, error)
	GetDashboard(ctx context.Context, userID int) (domain.Dashboard, error)
	DeleteUserTransactions(ctx context.Context, userID int) (int, error)
	GetMetrics(ctx context.Context) (postgres.Metrics, error)
}
