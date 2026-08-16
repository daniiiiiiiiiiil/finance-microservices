package grpc

import (
	"backend/internal/core/transport/grpc/interceptors"
	"backend/internal/features/finance/transport/grpc/proto"

	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *FinanceServer) GetTransaction(ctx context.Context, req *proto.GetTransactionRequest) (*proto.TransactionResponse, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	s.logger.Debug("gRPC GetTransaction", zap.Int32("id", req.Id), zap.Int("user_id", userID))

	if req.Id == 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	tx, err := s.service.GetTransaction(ctx, int(req.Id))
	if err != nil {
		s.logger.Error("get transaction", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	if tx.UserID != userID {
		isAdmin := interceptors.IsAdmin(ctx)
		if !isAdmin {
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		}
	}
	return convertFinanceToProto(tx), nil
}
