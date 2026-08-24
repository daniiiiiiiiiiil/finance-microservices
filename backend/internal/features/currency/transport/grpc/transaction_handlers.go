package grpc

import (
	"backend/pkg/grpcutil/interceptors"
	"backend/proto/currency/gen"

	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *CurrencyServer) GetTransactionUSD(ctx context.Context, req *gen.GetTransactionUSDRequest) (*gen.GetTransactionUSDResponse, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.NotFound, "User not found in context")
	}
	s.logger.Debug("gRPC GetTransactionUSD",
		zap.Int("user_id", userID),
		zap.Int64("transaction_id", req.Id),
	)

	tx, err := s.service.GetTransactionUSD(ctx, int(req.Id))
	if err != nil {
		s.logger.Error("get transaction usd", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &gen.GetTransactionUSDResponse{
		TransactionId:    int64(tx.TransactionID),
		AmountUsd:        tx.AmountUSD,
		OriginalAmount:   tx.OriginalAmount,
		OriginalCurrency: tx.OriginalCurrency,
		ConvertedAt:      timestamppb.New(tx.ConvertedAt),
	}, nil
}
