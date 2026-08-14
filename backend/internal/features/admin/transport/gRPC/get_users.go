package gRPC

import (
	"backend/internal/core/transport/grpc/interceptors"
	"backend/internal/features/admin/transport/gRPC/proto"
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *AdminServer) GetUsers(ctx context.Context, req *emptypb.Empty) (*proto.GetUsersResponse, error) {
	adminID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	if !interceptors.IsAdmin(ctx) {
		return nil, status.Error(codes.PermissionDenied, "admin access required")
	}

	s.logger.Debug("gRPC GetUsers", zap.Int("admin_id", adminID))

	users, err := s.service.GetUsers(ctx, 20, 0)
	if err != nil {
		s.logger.Error("failed to get users", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return convertUsersToProto(users, 20, 0), nil
}
