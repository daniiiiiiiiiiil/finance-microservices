package grpc

import (
	"backend/internal/core/transport/grpc/interceptors"

	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *UserServer) GetUser(ctx context.Context, req *GetUserRequest) (*UserResponse, error) {
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
