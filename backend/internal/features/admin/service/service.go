package service_admin

import (
	"backend/internal/core/cache"
	financeСlient "backend/internal/core/clients/finance"
	"backend/internal/core/clients/users"
	"backend/internal/core/kafka"
	"backend/internal/core/logger"
	"backend/internal/core/repository/postgres/pool"
	"context"
	"fmt"
)

////go:generate mockgen -destination=mocks/mock_admin_service.go -package=mocks -source=service.go AdminService
//type AdminRepository interface {
//	GetUsers(ctx context.Context, limit, offset int) ([]domain.User, int, error)
//	GetUser(ctx context.Context, id int) (domain.User, error)
//	DeleteUserTx(ctx context.Context, tx pool.Tx, id int) error
//	UpdateUserRoleTx(ctx context.Context, tx pool.Tx, id int, isAdmin bool) (domain.User, error)
//	GetMetrics(ctx context.Context) (Metrics, error)
//}

type Metrics struct {
	TotalUsers        int     `json:"total_users"`
	TotalTransactions int     `json:"total_transactions"`
	TotalBalance      float64 `json:"total_balance"`
}

type AdminService struct {
	//repo          AdminRepository
	pool          pool.Pool
	redis         cache.RedisInterface
	producer      *kafka.Producer
	userClient    *users.UsersClient
	financeClient financeСlient.FinanceClientInterface //интерфейс чтобы не было цикличности
	logger        *logger.Logger
}

func NewAdminService(
	//repo AdminRepository,
	//p pool.Pool,
	redis cache.RedisInterface,
	producer *kafka.Producer,
	userClient *users.UsersClient,
	financeClient financeСlient.FinanceClientInterface,
	logger *logger.Logger) *AdminService {
	return &AdminService{
		//repo:          repo,
		//pool:          p,
		redis:         redis,
		producer:      producer,
		userClient:    userClient,
		financeClient: financeClient,
		logger:        logger,
	}
}

func (s *AdminService) invalidateCache(ctx context.Context, userID int) {
	userKey := fmt.Sprintf("user:%d", userID)
	if err := s.redis.Delete(ctx, userKey); err != nil {
		fmt.Printf("failed to invalidate user cache: %v\n", err)
	}

	metricsKey := "admin:metrics"
	if err := s.redis.Delete(ctx, metricsKey); err != nil {
		fmt.Printf("failed to invalidate metrics cache: %v\n", err)
	}
}
