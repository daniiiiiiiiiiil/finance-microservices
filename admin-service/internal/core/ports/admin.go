package ports

import (
	"context"

	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/internal/core/domain"
)

type AdminServiceInterface interface {
	GetUsers(ctx context.Context, limit, offset int) ([]domain.User, error)
	GetUser(ctx context.Context, id int) (domain.User, error)
	DeleteUser(ctx context.Context, id int, adminID int) error
	UpdateUserRole(ctx context.Context, id int, isAdmin bool) (domain.User, error)
	GetMetrics(ctx context.Context) (Metrics, error)
	StartConsumer(ctx context.Context)
}

type Metrics struct {
	TotalUsers        int     `json:"total_users"`
	TotalTransactions int     `json:"total_transactions"`
	TotalBalance      float64 `json:"total_balance"`
}
