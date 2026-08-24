package grpc

import (
	"backend/pkg/grpcutil/interceptors"
	"backend/proto/admin/gen"

	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *AdminServer) UpdateUserRole(ctx context.Context, req *gen.UpdateRoleRequest) (*gen.AdminUserResponse, error) {
	adminID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	if !interceptors.IsAdmin(ctx) {
		return nil, status.Error(codes.PermissionDenied, "admin access required")
	}

	s.logger.Debug("gRPC UpdateUserRole", zap.Int("admin_id", adminID))

	user, err := s.service.UpdateUserRole(ctx, int(req.Id), req.IsAdmin)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return convertUserToProto(user), nil
}
