package gRPC

import (
	"backend/internal/core/transport/grpc/interceptors"
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *AdminServer) GetUser(ctx context.Context, req *GetUserRequest) (*AdminUserResponse, error) {
	adminID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	if !interceptors.IsAdmin(ctx) {
		return nil, status.Error(codes.PermissionDenied, "admin access required")
	}

	s.logger.Debug("gRPC GetUser", zap.Int32("id", req.Id), zap.Int("admin_id", adminID))

	user, err := s.service.GetUser(ctx, int(req.Id))
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return convertUserToProto(user), nil
}
