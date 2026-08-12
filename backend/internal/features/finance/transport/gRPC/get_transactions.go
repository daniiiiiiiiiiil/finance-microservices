package gRPC

import (
	"backend/internal/core/transport/grpc/interceptors"

	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *FinanceServer) GetTransactions(ctx context.Context, req *GetTransactionsRequest) (*GetTransactionsResponse, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	s.logger.Debug("gRPC GetTransactions", zap.Int("user_id", userID), zap.String("type", req.GetType()), zap.String("category", req.GetCategory()))

	transactions, err := s.service.GetTransactions(ctx, userID, nil, nil, nil, nil, int(req.Limit), int(req.Offset))
	if err != nil {
		s.logger.Error("get transactions", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	protoTxs := make([]*TransactionResponse, len(transactions))
	for i, tx := range transactions {
		protoTxs[i] = convertFinanceToProto(tx)
	}
	return &GetTransactionsResponse{
		Transactions: protoTxs,
		Total:        int32(len(transactions)),
	}, nil
}
