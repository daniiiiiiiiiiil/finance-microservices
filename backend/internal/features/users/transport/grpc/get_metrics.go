package grpc

import (
	"backend/internal/features/users/transport/grpc/proto"
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *UserServer) GetMetrics(ctx context.Context, req *emptypb.Empty) (*proto.MetricsResponse, error) {
	s.logger.Debug("gRPC GetMetrics")

	totalUsers, err := s.service.GetMetrics(ctx)
	if err != nil {
		s.logger.Error("get metrics failed", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.MetricsResponse{
		TotalUsers: int32(totalUsers),
	}, nil
}
