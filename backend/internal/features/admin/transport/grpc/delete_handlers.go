package grpc

import (
	"backend/pkg/grpcutil/interceptors"
	"backend/proto/admin/gen"

	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *AdminServer) DeleteUser(ctx context.Context, req *gen.DeleteUserRequest) (*emptypb.Empty, error) {
	adminID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	if !interceptors.IsAdmin(ctx) {
		return nil, status.Error(codes.PermissionDenied, "admin access required")
	}

	if int(req.Id) == adminID {
		return nil, status.Error(codes.PermissionDenied, "admin cannot delete themselves")
	}

	s.logger.Debug("gRPC DeleteUser", zap.Int32("id", req.Id), zap.Int("admin_id", adminID))

	err := s.service.DeleteUser(ctx, int(req.Id), adminID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}
