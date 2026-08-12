package grpc

import (
	"backend/internal/core/logger"
	service_user "backend/internal/features/users/service"

	"google.golang.org/grpc"
)

type UserServer struct {
	UnimplementedUserServiceServer
	service *service_user.UsersService
	logger  *logger.Logger
}

func NewUserServer(serviceUser *service_user.UsersService, logger *logger.Logger) *UserServer {
	return &UserServer{service: serviceUser, logger: logger}
}

func RegisterUserServer(grpcServer *grpc.Server, userServer *UserServer) {
	RegisterUserServiceServer(grpcServer, userServer)
}
