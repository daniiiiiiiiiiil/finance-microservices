package gRPC

import (
	"backend/internal/core/domain"
	"backend/internal/core/transport/grpc/interceptors"
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *FinanceServer) UpdateTransaction(ctx context.Context, req *UpdateTransactionRequest) (*TransactionResponse, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	s.logger.Debug("gRPC UpdateTransaction", zap.Int("user_id", userID))

	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if req.TypeTransaction == "" {
		return nil, status.Error(codes.InvalidArgument, "type transaction is required")
	}
	if req.Amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount is greater than zero")
	}
	if req.Category == "" {
		return nil, status.Error(codes.InvalidArgument, "category is required")
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
	transaction := domain.Finance{
		TypeTransaction: req.TypeTransaction,
		Amount:          req.Amount,
		Category:        req.Category,
		CreatedAt:       time.Now(),
		UserID:          userID,
	}
	updated, err := s.service.UpdateTransaction(ctx, transaction)
	if err != nil {
		s.logger.Error("update transaction", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	return convertFinanceToProto(updated), nil
}
