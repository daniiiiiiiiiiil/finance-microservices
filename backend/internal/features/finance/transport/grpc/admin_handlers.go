package grpc

import (
	"backend/pkg/grpcutil/interceptors"
	"backend/proto/finance/gen"

	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *FinanceServer) DeleteUserTransaction(ctx context.Context, req *gen.DeleteUserTransactionsRequest) (*gen.DeleteUserTransactionsResponse, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "user id not found in context")
	}
	s.logger.Debug("gRPC DeleteUserTransactions", zap.Int("user_id", userID))

	if req.UserId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	if int(req.UserId) != userID && !interceptors.IsAdmin(ctx) {
		return nil, status.Error(codes.PermissionDenied, "access denied")
	}

	count, err := s.service.DeleteUserTransactions(ctx, int(req.UserId))
	if err != nil {
		s.logger.Error("gRPC DeleteUserTransactions", zap.Error(err))
		return nil, status.Error(codes.Internal, "delete user transactions error")
	}
	return &gen.DeleteUserTransactionsResponse{
		Success:      true,
		DeletedCount: int32(count),
	}, nil
}

func (s *FinanceServer) GetMetrics(ctx context.Context, req *emptypb.Empty) (*gen.MetricsResponse, error) {
	s.logger.Debug("gRPC GetMetrics")

	metrics, err := s.service.GetMetrics(ctx)
	if err != nil {
		s.logger.Error("get metrics failed", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &gen.MetricsResponse{
		TotalTransactions: int32(metrics.TotalTransactions),
		TotalBalance:      metrics.TotalBalance,
	}, nil
}
