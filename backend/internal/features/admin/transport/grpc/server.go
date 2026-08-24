package grpc

import (
	"backend/internal/features/admin/service"
	"backend/pkg/logger"
	"backend/proto/admin/gen"

	"google.golang.org/grpc"
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
