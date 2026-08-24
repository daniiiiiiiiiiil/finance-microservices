package grpc

import (
	"backend/pkg/grpcutil/interceptors"
	"backend/proto/users/gen"

	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *UserServer) UpdateRole(ctx context.Context, req *gen.UpdateRoleRequest) (*gen.UserResponse, error) {
	adminID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	if !interceptors.IsAdmin(ctx) {
		return nil, status.Error(codes.PermissionDenied, "admin access required")
	}

	if int(req.Id) == adminID {
		return nil, status.Error(codes.PermissionDenied, "admin cannot change their own role")
	}

	s.logger.Debug("gRPC UpdateRole",
		zap.Int32("id", req.Id),
		zap.Bool("is_admin", req.IsAdmin),
		zap.Int("admin_id", adminID),
	)

	users, err := s.service.UpdateRole(ctx, int(req.Id), req.IsAdmin)
	if err != nil {
		s.logger.Error("failed to update role", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return convertUserToProto(users), nil
}
