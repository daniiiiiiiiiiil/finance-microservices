package grpc

import (
	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/internal/core/ports"
	"google.golang.org/grpc"

	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/pkg/logger"
	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/proto/admin/gen"
)

type AdminServer struct {
	gen.UnimplementedAdminServiceServer
	service ports.AdminServiceInterface
	logger  *logger.Logger
}

func NewAdminServer(service ports.AdminServiceInterface, logger *logger.Logger) *AdminServer {
	return &AdminServer{
		service: service,
		logger:  logger,
	}
}

func RegisterAdminServer(grpcServer *grpc.Server, server *AdminServer) {
	gen.RegisterAdminServiceServer(grpcServer, server)
}
