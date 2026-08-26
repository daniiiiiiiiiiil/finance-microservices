package grpc

import (
	"google.golang.org/grpc"

	service_user "github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/features/users/service"
	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/pkg/logger"
	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/proto/users/gen"
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
