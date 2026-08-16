package grpc

import (
	"backend/internal/core/transport/grpc/interceptors"
	"backend/internal/features/admin/transport/grpc/proto"
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *AdminServer) DeleteUser(ctx context.Context, req *proto.DeleteUserRequest) (*emptypb.Empty, error) {
	adminID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	if !interceptors.IsAdmin(ctx) {
		return nil, status.Error(codes.PermissionDenied, "admin access required")
	}

	s.logger.Debug("gRPC DeleteUser", zap.Int32("id", req.Id), zap.Int("admin_id", adminID))

	err := s.service.DeleteUser(ctx, int(req.Id))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}
