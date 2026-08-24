package grpc

import (
	service2 "backend/internal/features/finance/service"
	"backend/pkg/logger"
	"backend/proto/finance/gen"

	"google.golang.org/grpc"
)

type FinanceServer struct {
	gen.UnimplementedFinanceServiceServer
	service service2.FinanceServiceInterface
	logger  *logger.Logger
}

func NewFinanceServer(service service2.FinanceServiceInterface, logger *logger.Logger) *FinanceServer {
	return &FinanceServer{
		service: service,
		logger:  logger,
	}
}

func RegisterFinanceServer(grpcServer *grpc.Server, financeServer *FinanceServer) {
	gen.RegisterFinanceServiceServer(grpcServer, financeServer)
}
