package grpc

import (
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/ports"
	"google.golang.org/grpc"

	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/pkg/logger"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/proto/finance/gen"
)

type FinanceServer struct {
	gen.UnimplementedFinanceServiceServer
	service ports.FinanceServiceInterface
	logger  *logger.Logger
}

func NewFinanceServer(service ports.FinanceServiceInterface, logger *logger.Logger) *FinanceServer {
	return &FinanceServer{
		service: service,
		logger:  logger,
	}
}

func RegisterFinanceServer(grpcServer *grpc.Server, financeServer *FinanceServer) {
	gen.RegisterFinanceServiceServer(grpcServer, financeServer)
}
