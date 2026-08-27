package grpc

import (
	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/core/ports"
	"google.golang.org/grpc"

	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/pkg/logger"
	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/proto/auth/gen"
)

type AuthServer struct {
	gen.UnimplementedAuthServiceServer
	service ports.AuthServiceInterface
	logger  *logger.Logger
}

func NewAuthServer(service ports.AuthServiceInterface, logger *logger.Logger) *AuthServer {
	return &AuthServer{
		service: service,
		logger:  logger,
	}
}

func RegisterAuthServer(server *grpc.Server, auth *AuthServer) {
	gen.RegisterAuthServiceServer(server, auth)
}
