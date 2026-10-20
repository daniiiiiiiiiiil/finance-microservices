package grpc

import (
	"context"

	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/pkg/grpcutil/interceptors"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/proto/finance/gen"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *FinanceServer) GetAnalytics(ctx context.Context, req *gen.GetAnalyticsRequest) (*gen.AnalyticsResponse, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	s.logger.Debug("gRPC GetAnalytics",
		zap.Int("user_id", userID),
		zap.String("period", req.Period),
	)

	analytics, err := s.service.GetAnalytics(ctx, userID, req.Period)
	if err != nil {
		s.logger.Error("get analytics failed", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return convertAnalyticsToProto(analytics), nil
}
