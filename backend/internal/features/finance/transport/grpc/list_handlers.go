package grpc

import (
	"backend/pkg/grpcutil/interceptors"
	"backend/proto/finance/gen"
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *FinanceServer) GetTransactions(ctx context.Context, req *gen.GetTransactionsRequest) (*gen.GetTransactionsResponse, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	s.logger.Debug("gRPC GetTransactions",
		zap.Int("user_id", userID),
		zap.String("type", req.GetType()),
		zap.String("category", req.GetCategory()),
		zap.Int32("limit", req.Limit),
		zap.Int32("offset", req.Offset),
	)

	var transactionType, category *string
	if req.GetType() != "" {
		transactionType = req.Type
	}
	if req.GetCategory() != "" {
		category = req.Category
	}

	var from, to *time.Time
	if req.GetFrom() != "" {
		t, err := time.Parse("2006-01-02", req.GetFrom())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid 'from' date format (use YYYY-MM-DD)")
		}
		from = &t
	}
	if req.GetTo() != "" {
		t, err := time.Parse("2006-01-02", req.GetTo())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid 'to' date format (use YYYY-MM-DD)")
		}
		t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		to = &t
	}

	total, err := s.service.GetTransactionsCount(ctx, userID, transactionType, category, from, to)
	if err != nil {
		s.logger.Error("get transactions count", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	transactions, err := s.service.GetTransactions(ctx, userID, transactionType, category, from, to, int(req.Limit), int(req.Offset))
	if err != nil {
		s.logger.Error("get transactions", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoTxs := make([]*gen.TransactionResponse, len(transactions))
	for i, tx := range transactions {
		protoTxs[i] = convertFinanceToProto(tx)
	}

	return &gen.GetTransactionsResponse{
		Transactions: protoTxs,
		Total:        int32(total),
		Limit:        req.Limit,
		Offset:       req.Offset,
	}, nil
}

func (s *FinanceServer) GetCategories(ctx context.Context, req *gen.GetCategoriesRequest) (*gen.GetCategoriesResponse, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	s.logger.Debug("gRPC GetCategories", zap.Int("user_id", userID))

	categories, err := s.service.GetCategories(ctx, userID)
	if err != nil {
		s.logger.Error("get categories", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &gen.GetCategoriesResponse{
		Categories: categories,
	}, nil
}
