package grpc

import (
	service_currency "github.com/daniiiiiiiiiiil/finance-microservices/currency-service/internal/features/currency/service"
	"github.com/daniiiiiiiiiiil/finance-microservices/currency-service/pkg/logger"
	"github.com/daniiiiiiiiiiil/finance-microservices/currency-service/proto/currency/gen"
	"google.golang.org/grpc"
)

type CurrencyServer struct {
	gen.UnimplementedCurrencyServiceServer
	service service_currency.CurrencyServiceInterface
	logger  *logger.Logger
}

func NewCurrencyServer(service service_currency.CurrencyServiceInterface, logger *logger.Logger) *CurrencyServer {
	return &CurrencyServer{
		service: service,
		logger:  logger,
	}
}

func RegisterCurrencyServer(grpcServer *grpc.Server, currencyServer *CurrencyServer) {
	gen.RegisterCurrencyServiceServer(grpcServer, currencyServer)
}
