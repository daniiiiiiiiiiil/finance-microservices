package grpc

import (
	"backend/internal/core/logger"
	service_user "backend/internal/features/users/service"
	"backend/internal/features/users/transport/grpc/proto"

	"google.golang.org/grpc"
)

type UserServer struct {
	proto.UnimplementedUserServiceServer
	service *service_user.UsersService
	logger  *logger.Logger
}

func NewUserServer(serviceUser *service_user.UsersService, logger *logger.Logger) *UserServer {
	return &UserServer{service: serviceUser, logger: logger}
}

func RegisterUserServer(grpcServer *grpc.Server, userServer *UserServer) {
	proto.RegisterUserServiceServer(grpcServer, userServer)
}
