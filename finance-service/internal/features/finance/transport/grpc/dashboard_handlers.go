package grpc

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/pkg/grpcutil/interceptors"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/proto/finance/gen"
)

func (s *FinanceServer) GetDashboard(ctx context.Context, req *gen.GetDashboardRequest) (*gen.DashboardResponse, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	s.logger.Debug("gRPC GetDashboard", zap.Int("user_id", userID))
	dashboard, err := s.service.GetDashboard(ctx, userID)
	if err != nil {
		s.logger.Error("get dashboard", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	return convertDashboardToProto(dashboard), nil
}
