package service_admin

import (
	financeClient "github.com/daniiiiiiiiiiil/finance-microservices/admin-service/internal/core/clients/finance"
	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/pkg/logger"
)

var _ ports.AdminServiceInterface = (*AdminService)(nil)

type AdminService struct {
	redis          ports.RedisInterface
	eventPublisher ports.EventPublisherInterface
	userClient     ports.UsersClientInterface
	financeClient  financeClient.FinanceClientInterface
	logger         *logger.Logger
}

func NewAdminService(
	redis ports.RedisInterface,
	eventPublisher ports.EventPublisherInterface,
	userClient ports.UsersClientInterface,
	financeClient financeClient.FinanceClientInterface,
	logger *logger.Logger,
) *AdminService {
	return &AdminService{
		redis:          redis,
		eventPublisher: eventPublisher,
		userClient:     userClient,
		financeClient:  financeClient,
		logger:         logger,
	}
}
