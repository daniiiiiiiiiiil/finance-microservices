package grpc

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/pkg/grpcutil/interceptors"
	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/proto/admin/gen"
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
