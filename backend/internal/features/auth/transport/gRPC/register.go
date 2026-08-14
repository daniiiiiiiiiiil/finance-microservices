package gRPC

import (
	"backend/internal/core/domain"
	service "backend/internal/features/auth/service"
	"backend/internal/features/auth/transport/gRPC/proto"
	"backend/internal/features/auth/transport/http/dto"

	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *AuthServer) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.AuthResponse, error) {
	if req.Email == "" || len(req.Email) <= 0 {
		return nil, status.Error(codes.InvalidArgument, "Email is required")
	}
	if req.Password == "" || len(req.Password) <= 0 {
		return nil, status.Error(codes.InvalidArgument, "Password is required")
	}
	if len(req.Password) < 8 {
		return nil, status.Error(codes.InvalidArgument, "Password must be at least 8 characters")
	}
	s.logger.Debug("Auth Register", zap.String("email", req.Email))

	registerReq := dto.RegisterRequest{
		FullName:    req.FullName,
		Email:       req.Email,
		Password:    req.Password,
		PhoneNumber: req.PhoneNumber,
		IsAdmin:     req.IsAdmin,
	}
	token, user, err := s.service.Register(ctx, registerReq)
	if err != nil {
		s.logger.Error("register failed", zap.Error(err))
		if err == service.ErrUserAlreadyExists {
			return nil, status.Error(codes.AlreadyExists, "User already exists")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	userDomain := domain.User{
		ID:          user.ID,
		FullName:    user.FullName,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		IsAdmin:     user.IsAdmin,
	}
	return &proto.AuthResponse{
		Token: token,
		User:  convertUserToProto(userDomain),
	}, nil
}
