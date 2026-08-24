package grpc

import (
	"backend/pkg/grpcutil/interceptors"
	"backend/proto/currency/gen"

	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *CurrencyServer) GetRates(ctx context.Context, req *gen.GetRatesRequest) (*gen.GetRatesResponse, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.NotFound, "User not found")
	}
	s.logger.Debug("gRPC GetRates",
		zap.Int("user_id", userID),
		zap.String("base", req.Base),
	)

	rates, err := s.service.GetRates(ctx, req.Base)
	if err != nil {
		s.logger.Error("gRPC GetRates error", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	return convertRateToProto(rates), nil
}
