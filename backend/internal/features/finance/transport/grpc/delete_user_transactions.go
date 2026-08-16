package grpc

import (
	"backend/internal/core/transport/grpc/interceptors"
	"backend/internal/features/finance/transport/grpc/proto"

	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *FinanceServer) DeleteUserTransactions(ctx context.Context, req *proto.DeleteUserTransactionsRequest) (*proto.DeleteUserTransactionsResponse, error) {
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
	return &proto.DeleteUserTransactionsResponse{
		Success:      true,
		DeletedCount: int32(count),
	}, nil
}
