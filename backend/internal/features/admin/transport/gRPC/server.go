package gRPC

import (
	"backend/internal/core/logger"
	service_admin "backend/internal/features/admin/service"

	"google.golang.org/grpc"
)

type AdminServer struct {
	UnimplementedAdminServiceServer
	service *service_admin.AdminService
	logger  *logger.Logger
}

func NewAdminServer(service *service_admin.AdminService, logger *logger.Logger) *AdminServer {
	return &AdminServer{
		service: service,
		logger:  logger,
	}
}

func RegisterAdminServer(grpcServer *grpc.Server, server *AdminServer) {
	RegisterAdminServiceServer(grpcServer, server)
}
