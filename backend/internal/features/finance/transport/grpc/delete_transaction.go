package grpc

import (
	"backend/internal/core/transport/grpc/interceptors"
	"backend/internal/features/finance/transport/grpc/proto"

	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *FinanceServer) DeleteTransaction(ctx context.Context, req *proto.DeleteTransactionRequest) (*emptypb.Empty, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	s.logger.Debug("gRPC DeleteTransaction", zap.Int("user_id", userID))

	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	if int(req.Id) != userID && !interceptors.IsAdmin(ctx) {
		return nil, status.Error(codes.PermissionDenied, "access denied")
	}

	existing, err := s.service.GetTransaction(ctx, int(req.Id))
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if existing.UserID != userID {
		isAdmin := interceptors.IsAdmin(ctx)
		if !isAdmin {
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		}
	}
	err = s.service.DeleteTransaction(ctx, int(req.Id))
	if err != nil {
		s.logger.Error("delete transaction", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}
