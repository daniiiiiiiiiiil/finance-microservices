package grpc

import (
	"backend/internal/core/domain"
	"backend/internal/features/auth/transport/grpc/proto"
	http_auth "backend/internal/features/auth/transport/http/dto"
	"fmt"
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (s *AuthServer) Login(ctx context.Context, req *proto.LoginRequest) (*proto.AuthResponse, error) {
	if req.Email == "" || len(req.Email) <= 0 {
		return nil, status.Error(codes.InvalidArgument, "Email is required")
	}
	if req.Password == "" || len(req.Password) <= 0 {
		return nil, status.Error(codes.InvalidArgument, "Password is required")
	}

	s.logger.Debug("Auth Login", zap.String("email", req.Email))

	ip := "unknown"
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if clientIP := md.Get("x-forwarded-for"); len(clientIP) > 0 {
			ip = clientIP[0]
		}
	}
	key := fmt.Sprintf("rate:login:%s", ip)
	allowed, err := s.service.RateLimitCheck(ctx, key, 5, 1*time.Minute)
	if err != nil {
		s.logger.Warn("rate limit check failed", zap.Error(err))
	}
	if !allowed {
		return nil, status.Error(codes.ResourceExhausted, "too many login attempts, try again later")
	}

	loginReq := http_auth.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	token, user, err := s.service.Login(ctx, loginReq)
	if err != nil {
		s.logger.Error("login failed", zap.Error(err))
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
