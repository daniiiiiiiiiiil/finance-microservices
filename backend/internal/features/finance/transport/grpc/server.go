package grpc

import (
	"backend/internal/core/logger"
	service_finance "backend/internal/features/finance/service"
	"backend/internal/features/finance/transport/grpc/proto"

	"google.golang.org/grpc"
)

type FinanceServer struct {
	proto.UnimplementedFinanceServiceServer
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
	proto.RegisterFinanceServiceServer(grpcServer, financeServer)
}
