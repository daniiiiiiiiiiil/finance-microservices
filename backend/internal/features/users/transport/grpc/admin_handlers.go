package grpc

import (
	"backend/proto/users/gen"

	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *UserServer) AdminExists(ctx context.Context, req *emptypb.Empty) (*gen.AdminExistsResponse, error) {
	exists, err := s.service.AdminExists(ctx)
	if err != nil {
		return nil, err
	}
	return &gen.AdminExistsResponse{Exists: exists}, nil
}

func (s *UserServer) GetMetrics(ctx context.Context, req *emptypb.Empty) (*gen.MetricsResponse, error) {
	s.logger.Debug("gRPC GetMetrics")

	totalUsers, err := s.service.GetMetrics(ctx)
	if err != nil {
		s.logger.Error("get metrics failed", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &gen.MetricsResponse{
		TotalUsers: int32(totalUsers),
	}, nil
}
