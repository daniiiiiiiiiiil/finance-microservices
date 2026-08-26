package grpc

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	service_user "github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/features/users/service"
	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/proto/users/gen"
)

func (s *UserServer) CreateProfile(ctx context.Context, req *gen.CreateProfileRequest) (*gen.UserResponse, error) {
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

func (s *UserServer) GetUserByEmail(ctx context.Context, req *gen.GetUserByEmailRequest) (*gen.UserResponse, error) {
	user, err := s.service.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return convertUserToProto(user), nil
}
