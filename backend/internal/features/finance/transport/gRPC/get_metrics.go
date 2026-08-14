package gRPC

import (
	"backend/internal/features/finance/transport/gRPC/proto"
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *FinanceServer) GetMetrics(ctx context.Context, req *emptypb.Empty) (*proto.MetricsResponse, error) {
	s.logger.Debug("gRPC GetMetrics")

	metrics, err := s.service.GetMetrics(ctx)
	if err != nil {
		s.logger.Error("get metrics failed", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.MetricsResponse{
		TotalTransactions: int32(metrics.TotalTransactions),
		TotalBalance:      metrics.TotalBalance,
	}, nil
}
