package gRPC

import (
	"backend/internal/core/transport/grpc/interceptors"
	"backend/internal/features/finance/transport/gRPC/proto"

	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *FinanceServer) GetCategories(ctx context.Context, req *proto.GetCategoriesRequest) (*proto.GetCategoriesResponse, error) {
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
	return &proto.GetCategoriesResponse{
		Categories: categories,
	}, nil
}
