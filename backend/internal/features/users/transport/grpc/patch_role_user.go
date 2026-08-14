package grpc

import (
	"backend/internal/core/transport/grpc/interceptors"
	"backend/internal/features/users/transport/grpc/proto"

	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *UserServer) PatchUserRole(ctx context.Context, req *proto.UpdateRoleRequest) (*proto.UserResponse, error) {
	UserID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "UserID not found in context")
	}
	if UserID != int(req.Id) {
		isAdmin := interceptors.IsAdmin(ctx)
		if !isAdmin {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
	}
	s.logger.Debug("gRPC patch user role", zap.Int32("id", req.Id))

	users, err := s.service.UpdateRoleUsers(ctx, int(req.Id), req.IsAdmin)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return convertUserToProto(users), nil
}
