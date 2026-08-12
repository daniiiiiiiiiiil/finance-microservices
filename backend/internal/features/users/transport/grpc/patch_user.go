package grpc

import (
	"backend/internal/core/transport/grpc/interceptors"

	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *UserServer) PatchUser(ctx context.Context, req *PatchUserRequest) (*UserResponse, error) {
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
