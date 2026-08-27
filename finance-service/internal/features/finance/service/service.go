package service

import (
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/repository/postgres/pool"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/pkg/logger"
)

var _ FinanceServiceInterface = (*FinanceService)(nil)

type FinanceService struct {
	repo           ports.FinanceRepositoryInterface
	pool           pool.Pool
	redis          ports.RedisInterface
	eventPublisher ports.EventPublisherInterface
	logger         *logger.Logger
}

func NewFinanceService(
	repo ports.FinanceRepositoryInterface,
	pool pool.Pool,
	redis ports.RedisInterface,
	eventPublisher ports.EventPublisherInterface,
	logger *logger.Logger,
) *FinanceService {
	return &FinanceService{
		repo:           repo,
		pool:           pool,
		redis:          redis,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}
