package grpc

import (
	"backend/internal/core/domain"
	service "backend/internal/features/auth/service"
	"backend/pkg/grpcutil/interceptors"
	"backend/proto/auth/gen"
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (s *AuthServer) Register(ctx context.Context, req *gen.RegisterRequest) (*gen.AuthResponse, error) {
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

	registerReq := service.RegisterRequest{
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
	return &gen.AuthResponse{
		Token: token,
		User:  convertUserToProto(userDomain),
	}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *gen.LoginRequest) (*gen.AuthResponse, error) {
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

	loginReq := service.LoginRequest{
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

	return &gen.AuthResponse{
		Token: token,
		User:  convertUserToProto(userDomain),
	}, nil
}

func (s *AuthServer) Logout(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}
	s.logger.Debug("Logout", zap.Int("user_id", userID))

	md, _ := metadata.FromIncomingContext(ctx)
	auth := md.Get("authorization")
	if len(auth) > 0 {
		token := strings.TrimPrefix(auth[0], "Bearer ")
		if err := s.service.AddToBlacklist(ctx, token, 24*time.Hour); err != nil {
			return nil, status.Error(codes.Unauthenticated, "unauthenticated")
		}
	}
	return &empty.Empty{}, nil
}
