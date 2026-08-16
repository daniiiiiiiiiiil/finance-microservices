package grpc

import (
	"backend/internal/core/logger"
	service_admin "backend/internal/features/admin/service"
	"backend/internal/features/admin/transport/grpc/proto"

	"google.golang.org/grpc"
)

type AdminServer struct {
	proto.UnimplementedAdminServiceServer
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
	proto.RegisterAdminServiceServer(grpcServer, server)
}
