package grpc

import (
	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/pkg/logger"
	aipb "github.com/daniiiiiiiiiiil/finance-microservices/ai-service/proto/gen/v1"
	"google.golang.org/grpc"
)

type Server struct {
	aipb.UnimplementedAIServiceServer
	service ports.PlanServiceInterface
	logger  *logger.Logger
}

func NewServer(service ports.PlanServiceInterface, log *logger.Logger) *Server {
	return &Server{service: service, logger: log}
}

func Register(grpcServer *grpc.Server, srv *Server) {
	aipb.RegisterAIServiceServer(grpcServer, srv)
}
