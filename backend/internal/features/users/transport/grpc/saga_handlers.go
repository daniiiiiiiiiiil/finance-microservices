package grpc

import (
	"backend/pkg/grpcutil/interceptors"
	"backend/proto/users/gen"

	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *UserServer) MarkDeleting(ctx context.Context, req *gen.MarkDeletingRequest) (*emptypb.Empty, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	if !interceptors.IsAdmin(ctx) {
		return nil, status.Error(codes.PermissionDenied, "admin access required")
	}

	s.logger.Debug("gRPC MarkDeleting",
		zap.Int32("id", req.Id),
		zap.Int("admin_id", userID),
	)

	if err := s.service.MarkDeleting(ctx, int(req.Id)); err != nil {
		s.logger.Error("failed to mark deleting", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s *UserServer) FinalizeDelete(ctx context.Context, req *gen.FinalizeDeleteRequest) (*emptypb.Empty, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	if !interceptors.IsAdmin(ctx) {
		return nil, status.Error(codes.PermissionDenied, "admin access required")
	}

	s.logger.Debug("gRPC FinalizeDelete",
		zap.Int32("id", req.Id),
		zap.Int("admin_id", userID),
	)

	if err := s.service.FinalizeDelete(ctx, int(req.Id)); err != nil {
		s.logger.Error("failed to finalize delete", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s *UserServer) RestoreUser(ctx context.Context, req *gen.RestoreUserRequest) (*emptypb.Empty, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	if !interceptors.IsAdmin(ctx) {
		return nil, status.Error(codes.PermissionDenied, "admin access required")
	}

	s.logger.Debug("gRPC RestoreUser",
		zap.Int32("id", req.Id),
		zap.Int("admin_id", userID),
	)

	if err := s.service.RestoreUser(ctx, int(req.Id)); err != nil {
		s.logger.Error("failed to restore user", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}
