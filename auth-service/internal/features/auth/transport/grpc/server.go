package grpc

import (
	"google.golang.org/grpc"

	service_auth "github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/features/auth/service"
	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/pkg/logger"
	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/proto/auth/gen"
)

type AuthServer struct {
	gen.UnimplementedAuthServiceServer
	service service_auth.AuthServiceInterface
	logger  *logger.Logger
}

func NewAuthServer(service service_auth.AuthServiceInterface, logger *logger.Logger) *AuthServer {
	return &AuthServer{
		service: service,
		logger:  logger,
	}
}

func RegisterAuthServer(server *grpc.Server, auth *AuthServer) {
	gen.RegisterAuthServiceServer(server, auth)
}
