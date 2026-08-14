package gRPC

import (
	"backend/internal/core/logger"
	service "backend/internal/features/auth/service"
	"backend/internal/features/auth/transport/gRPC/proto"

	"google.golang.org/grpc"
)

type AuthServer struct {
	proto.UnimplementedAuthServiceServer
	service *service.AuthService
	logger  *logger.Logger
}

func NewAuthServer(service *service.AuthService, logger *logger.Logger) *AuthServer {
	return &AuthServer{
		service: service,
		logger:  logger,
	}
}

func RegisterAuthServer(server *grpc.Server, auth *AuthServer) {
	proto.RegisterAuthServiceServer(server, auth)
}
