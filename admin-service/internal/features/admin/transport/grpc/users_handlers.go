package grpc

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/pkg/grpcutil/interceptors"
	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/proto/admin/gen"
)

func (s *AdminServer) GetUser(ctx context.Context, req *gen.GetUserRequest) (*gen.AdminUserResponse, error) {
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

func (s *AdminServer) GetUsers(ctx context.Context, req *emptypb.Empty) (*gen.GetUsersResponse, error) {
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
