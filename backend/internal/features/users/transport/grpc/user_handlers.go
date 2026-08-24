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

func (s *UserServer) GetUser(ctx context.Context, req *gen.GetUserRequest) (*gen.UserResponse, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	if int(req.Id) != userID {
		isAdmin := interceptors.IsAdmin(ctx)
		if !isAdmin {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
	}

	s.logger.Debug("gRPC GetUser", zap.Int32("id", req.Id))

	user, err := s.service.GetUser(ctx, int(req.Id))
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return convertUserToProto(user), nil
}

func (s *UserServer) PatchUser(ctx context.Context, req *gen.PatchUserRequest) (*gen.UserResponse, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	if int(req.Id) != userID {
		isAdmin := interceptors.IsAdmin(ctx)
		if !isAdmin {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
	}

	s.logger.Debug("gRPC PatchUser", zap.Int32("id", req.Id))

	patch := convertPatchProtoToDomain(req.Data)

	user, err := s.service.PatchUser(ctx, int(req.Id), patch)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return convertUserToProto(user), nil
}

func (s *UserServer) DeleteUser(ctx context.Context, req *gen.DeleteUserRequest) (*emptypb.Empty, error) {
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	if int(req.Id) != userID {
		isAdmin := interceptors.IsAdmin(ctx)
		if !isAdmin {
			return nil, status.Error(codes.PermissionDenied, "access denied")
		}
	}

	s.logger.Debug("gRPC DeleteUser", zap.Int32("id", req.Id))

	err := s.service.DeleteUser(ctx, int(req.Id))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}
