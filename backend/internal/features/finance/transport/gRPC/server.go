package gRPC

import (
	"backend/internal/core/logger"
	service_finance "backend/internal/features/finance/service"

	"google.golang.org/grpc"
)

type FinanceServer struct {
	UnimplementedFinanceServiceServer
	service *service_finance.FinanceService
	logger  *logger.Logger
}

func NewFinanceServer(service *service_finance.FinanceService, logger *logger.Logger) *FinanceServer {
	return &FinanceServer{
		service: service,
		logger:  logger,
	}
}

func RegisterFinanceServer(grpcServer *grpc.Server, financeServer *FinanceServer) {
	RegisterFinanceServiceServer(grpcServer, financeServer)
}
