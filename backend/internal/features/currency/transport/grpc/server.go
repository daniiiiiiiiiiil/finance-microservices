package grpc

import (
	"backend/internal/core/logger"
	service_currency "backend/internal/features/currency/service"
	"backend/internal/features/currency/transport/grpc/proto"

	"google.golang.org/grpc"
)

type CurrencyServer struct {
	proto.UnimplementedCurrencyServiceServer
	service *service_currency.CurrencyService
	logger  *logger.Logger
}

func NewCurrencyServer(service *service_currency.CurrencyService, logger *logger.Logger) *CurrencyServer {
	return &CurrencyServer{
		service: service,
		logger:  logger,
	}
}

func RegisterCurrencyServer(grpcServer *grpc.Server, currencyServer *CurrencyServer) {
	proto.RegisterCurrencyServiceServer(grpcServer, currencyServer)
}
