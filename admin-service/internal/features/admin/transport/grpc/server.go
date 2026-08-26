package grpc

import (
	"google.golang.org/grpc"

	service_admin "github.com/daniiiiiiiiiiil/finance-microservices/admin-service/internal/features/admin/service"
	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/pkg/logger"
	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/proto/admin/gen"
)

type AdminServer struct {
	gen.UnimplementedAdminServiceServer
	service service_admin.AdminServiceInterface
	logger  *logger.Logger
}

func NewAdminServer(service service_admin.AdminServiceInterface, logger *logger.Logger) *AdminServer {
	return &AdminServer{
		service: service,
		logger:  logger,
	}
}

func RegisterAdminServer(grpcServer *grpc.Server, server *AdminServer) {
	gen.RegisterAdminServiceServer(grpcServer, server)
}
