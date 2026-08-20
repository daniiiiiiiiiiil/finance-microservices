package grpc

import (
	service_user "backend/internal/features/users/service"
	"backend/internal/features/users/transport/grpc/proto"

	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *UserServer) CreateProfile(ctx context.Context, req *proto.CreateProfileRequest) (*proto.UserResponse, error) {
	s.logger.Debug("gRPC CreateProfile", zap.String("email", req.Email))

	var phoneNumber *string
	if req.PhoneNumber != "" {
		phoneNumber = &req.PhoneNumber
	}

	user, err := s.service.CreateProfile(ctx, &service_user.CreateProfileRequest{
		Email:        req.Email,
		FullName:     req.FullName,
		PhoneNumber:  phoneNumber,
		IsAdmin:      req.IsAdmin,
		PasswordHash: req.PasswordHash,
	})
	if err != nil {
		s.logger.Error("create profile failed", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return convertUserToProto(user), nil
}
