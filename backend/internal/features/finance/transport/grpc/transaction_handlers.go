package grpc

import (
	"backend/internal/core/domain"
	"backend/pkg/grpcutil/interceptors"
	"backend/proto/finance/gen"
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *FinanceServer) CreateTransaction(ctx context.Context, req *gen.CreateTransactionRequest) (*gen.TransactionResponse, error) {
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

func (s *FinanceServer) GetTransaction(ctx context.Context, req *gen.GetTransactionRequest) (*gen.TransactionResponse, error) {
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

func (s *FinanceServer) UpdateTransaction(ctx context.Context, req *gen.UpdateTransactionRequest) (*gen.TransactionResponse, error) {
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
		ID:              int(req.Id),
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

func (s *FinanceServer) DeleteTransaction(ctx context.Context, req *gen.DeleteTransactionRequest) (*emptypb.Empty, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	s.logger.Debug("gRPC DeleteTransaction", zap.Int("user_id", userID))

	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	existing, err := s.service.GetTransaction(ctx, int(req.Id))
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if existing.UserID != userID && !interceptors.IsAdmin(ctx) {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	err = s.service.DeleteTransaction(ctx, int(req.Id))
	if err != nil {
		s.logger.Error("delete transaction", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}
