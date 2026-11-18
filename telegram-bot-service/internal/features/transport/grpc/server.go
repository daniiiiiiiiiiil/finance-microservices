package grpc

import (
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/logger"
	telegrampb "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/proto/telegram/gen"
	"google.golang.org/grpc"
)

type Server struct {
	telegrampb.UnimplementedTelegramServiceServer
	accounts ports.AccountServiceInterface
	logger   *logger.Logger
}

func NewServer(accounts ports.AccountServiceInterface, logger *logger.Logger) *Server {
	return &Server{accounts: accounts, logger: logger}
}

func RegisterTelegramServer(grpcServer *grpc.Server, server *Server) {
	telegrampb.RegisterTelegramServiceServer(grpcServer, server)
}
