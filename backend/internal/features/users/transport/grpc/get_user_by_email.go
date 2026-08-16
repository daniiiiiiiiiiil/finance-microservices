package grpc

import (
	"backend/internal/features/users/transport/grpc/proto"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *UserServer) GetUserByEmail(ctx context.Context, req *proto.GetUserByEmailRequest) (*proto.UserResponse, error) {
	user, err := s.service.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return convertUserToProto(user), nil
}
