package grpc

import (
	"backend/internal/features/auth/service"
	"backend/pkg/logger"
	"backend/proto/auth/gen"

	"google.golang.org/grpc"
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
