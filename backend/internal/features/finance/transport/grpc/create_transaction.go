package grpc

import (
	"backend/internal/core/domain"
	"backend/internal/core/transport/grpc/interceptors"
	"backend/internal/features/finance/transport/grpc/proto"
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *FinanceServer) CreateTransaction(ctx context.Context, req *proto.CreateTransactionRequest) (*proto.TransactionResponse, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	s.logger.Debug("gRPC CreateTransaction",
		zap.Int("user_id", userID),
		zap.String("type", req.TypeTransaction),
		zap.Float64("amount", req.Amount),
	)

	if req.TypeTransaction == "" {
		return nil, status.Error(codes.InvalidArgument, "type transaction is required")
	}
	if req.Amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be greater than zero")
	}
	if req.Category == "" {
		return nil, status.Error(codes.InvalidArgument, "category is required")
	}
	transaction := domain.Finance{
		TypeTransaction: req.TypeTransaction,
		Amount:          req.Amount,
		Category:        req.Category,
		CreatedAt:       time.Now(),
		UserID:          userID,
	}

	created, err := s.service.CreateTransaction(ctx, transaction)
	if err != nil {
		s.logger.Error("create transaction", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	return convertFinanceToProto(created), nil
}
