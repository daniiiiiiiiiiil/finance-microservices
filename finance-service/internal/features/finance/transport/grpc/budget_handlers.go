package grpc

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/pkg/grpcutil/interceptors"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/proto/finance/gen"
)

func (s *FinanceServer) CreateBudget(ctx context.Context, req *gen.CreateBudgetRequest) (*gen.CreateBudgetResponse, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	s.logger.Debug("gRPC CreateBudget",
		zap.Int("user_id", userID),
		zap.String("category", req.Category),
		zap.Float64("amount_limit", req.AmountLimit),
		zap.String("period", req.Period),
	)

	if req.Category == "" {
		return nil, status.Error(codes.InvalidArgument, "category is required")
	}
	if req.AmountLimit <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount_limit must be greater than zero")
	}
	switch domain.BudgetPeriod(req.Period) {
	case domain.BudgetPeriodDay, domain.BudgetPeriodWeek, domain.BudgetPeriodMonth, domain.BudgetPeriodYear:
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid period: expected day, week, month, or year")
	}

	b := domain.Budget{
		UserID:      userID,
		Category:    req.Category,
		AmountLimit: req.AmountLimit,
		Period:      domain.BudgetPeriod(req.Period),
	}

	created, err := s.service.CreateBudget(ctx, b)
	if err != nil {
		s.logger.Error("create budget", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &gen.CreateBudgetResponse{
		Budget: convertBudgetToProto(created),
	}, nil
}

func (s *FinanceServer) ListBudgets(ctx context.Context, req *gen.ListBudgetsRequest) (*gen.ListBudgetsResponse, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	s.logger.Debug("gRPC ListBudgets", zap.Int("user_id", userID))

	statuses, err := s.service.ListBudgetsWithStatus(ctx, userID)
	if err != nil {
		s.logger.Error("list budgets", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	out := make([]*gen.BudgetStatus, 0, len(statuses))
	for _, st := range statuses {
		out = append(out, convertBudgetStatusToProto(st))
	}
	return &gen.ListBudgetsResponse{Budgets: out}, nil
}

func (s *FinanceServer) UpdateBudget(ctx context.Context, req *gen.UpdateBudgetRequest) (*gen.UpdateBudgetResponse, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	s.logger.Debug("gRPC UpdateBudget",
		zap.Int("user_id", userID),
		zap.Int32("id", req.Id),
	)

	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if req.Category == "" {
		return nil, status.Error(codes.InvalidArgument, "category is required")
	}
	if req.AmountLimit <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount_limit must be greater than zero")
	}
	switch domain.BudgetPeriod(req.Period) {
	case domain.BudgetPeriodDay, domain.BudgetPeriodWeek, domain.BudgetPeriodMonth, domain.BudgetPeriodYear:
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid period: expected day, week, month, or year")
	}

	existing, err := s.service.GetBudget(ctx, int(req.Id))
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if existing.UserID != userID && !interceptors.IsAdmin(ctx) {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	b := domain.Budget{
		ID:          int(req.Id),
		UserID:      existing.UserID,
		Category:    req.Category,
		AmountLimit: req.AmountLimit,
		Period:      domain.BudgetPeriod(req.Period),
	}

	updated, err := s.service.UpdateBudget(ctx, b)
	if err != nil {
		s.logger.Error("update budget", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &gen.UpdateBudgetResponse{
		Budget: convertBudgetToProto(updated),
	}, nil
}

func (s *FinanceServer) DeleteBudget(ctx context.Context, req *gen.DeleteBudgetRequest) (*emptypb.Empty, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	s.logger.Debug("gRPC DeleteBudget",
		zap.Int("user_id", userID),
		zap.Int32("id", req.Id),
	)

	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	existing, err := s.service.GetBudget(ctx, int(req.Id))
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if existing.UserID != userID && !interceptors.IsAdmin(ctx) {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	if err := s.service.DeleteBudget(ctx, int(req.Id)); err != nil {
		s.logger.Error("delete budget", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}
