package gRPC

import (
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/pkg/logger"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/proto/shopping/gen"
	"google.golang.org/grpc"
)

type ShoppingListService struct {
	gen.UnimplementedShoppingServiceServer
	service ports.ShoppingServiceInterface
	logger  *logger.Logger
}

func NewShoppingListService(
	service ports.ShoppingServiceInterface,
	logger *logger.Logger,
) *ShoppingListService {
	return &ShoppingListService{
		service: service,
		logger:  logger,
	}
}

func RegisterUserServer(grpcServer *grpc.Server, userServer *ShoppingListService) {
	gen.RegisterShoppingServiceServer(grpcServer, userServer)
}
