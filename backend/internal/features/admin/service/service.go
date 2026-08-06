package service

import (
	"backend/internal/core/cache"
	"backend/internal/core/domain"
	"backend/internal/core/repository/postgres/pool"
	"context"
)

type AdminRepository interface {
	GetUsers(ctx context.Context, limit, offset int) ([]domain.User, int, error)
	GetUser(ctx context.Context, id int) (domain.User, error)
	DeleteUserTx(ctx context.Context, tx pool.Tx, id int) error
	UpdateUserRoleTx(ctx context.Context, tx pool.Tx, id int, isAdmin bool) (domain.User, error)
	GetMetrics(ctx context.Context) (Metrics, error)
}

type Metrics struct {
	TotalUsers        int     `json:"total_users"`
	TotalTransactions int     `json:"total_transactions"`
	TotalBalance      float64 `json:"total_balance"`
}

type AdminService struct {
	repo  AdminRepository
	pool  pool.Pool
	redis *cache.RedisClient
}

func NewAdminService(repo AdminRepository, p pool.Pool, redis *cache.RedisClient) *AdminService {
	return &AdminService{
		repo:  repo,
		pool:  p,
		redis: redis,
	}
}
