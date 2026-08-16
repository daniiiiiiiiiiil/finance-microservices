package grpc

import (
	"backend/internal/core/transport/grpc/interceptors"
	"backend/internal/features/admin/transport/grpc/proto"

	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *AdminServer) GetMetrics(ctx context.Context, req *emptypb.Empty) (*proto.MetricsResponse, error) {
	adminID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	if !interceptors.IsAdmin(ctx) {
		return nil, status.Error(codes.PermissionDenied, "admin access required")
	}

	s.logger.Debug("gRPC GetMetrics", zap.Int("admin_id", adminID))

	metrics, err := s.service.GetMetrics(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.MetricsResponse{
		TotalUsers:        int32(metrics.TotalUsers),
		TotalTransactions: int32(metrics.TotalTransactions),
		TotalBalance:      metrics.TotalBalance,
	}, nil
}
