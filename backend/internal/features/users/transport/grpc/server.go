package grpc

import (
	service_user "backend/internal/features/users/service"
	"backend/pkg/logger"
	"backend/proto/users/gen"

	"google.golang.org/grpc"
)

type UserServer struct {
	gen.UnimplementedUserServiceServer
	service service_user.UsersServiceInterface
	logger  *logger.Logger
}

func NewUserServer(service service_user.UsersServiceInterface, logger *logger.Logger) *UserServer {
	return &UserServer{
		service: service,
		logger:  logger,
	}
}

func RegisterUserServer(grpcServer *grpc.Server, userServer *UserServer) {
	gen.RegisterUserServiceServer(grpcServer, userServer)
}
