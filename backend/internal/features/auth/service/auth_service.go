package service_auth

import (
	"backend/internal/core/auth/jwt"
	http_auth "backend/internal/features/auth/transport/http/dto"
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type AuthService struct {
	userRepo   UserRepository
	jwtManager *jwt.JWTManager
}

type UserRepository interface {
	CreateUserWithAuth(ctx context.Context, fullName, email, passwordHash string, phoneNumber *string, isAdmin bool) (int, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	AdminExists(ctx context.Context) (bool, error)
}

type User struct {
	ID           int
	FullName     string
	Email        string
	PasswordHash string
	PhoneNumber  *string
	IsAdmin      bool
}

func NewAuthService(userRepo UserRepository, jwtManager *jwt.JWTManager) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

func (s *AuthService) Register(ctx context.Context, req http_auth.RegisterRequest) (string, *http_auth.UserResponse, error) {
	_, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return "", nil, ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", nil, fmt.Errorf("hash password: %w", err)
	}

	var phoneNumber *string
	if req.PhoneNumber != "" {
		phoneNumber = &req.PhoneNumber
	}

	userID, err := s.userRepo.CreateUserWithAuth(ctx, req.FullName, req.Email, string(hashedPassword), phoneNumber, req.IsAdmin)
	if err != nil {
		return "", nil, fmt.Errorf("create user: %w", err)
	}

	token, err := s.jwtManager.Generate(userID)
	if err != nil {
		return "", nil, fmt.Errorf("generate token: %w", err)
	}

	userResponse := &http_auth.UserResponse{
		ID:          userID,
		FullName:    req.FullName,
		Email:       req.Email,
		PhoneNumber: phoneNumber,
		IsAdmin:     req.IsAdmin,
	}

	return token, userResponse, nil
}

func (s *AuthService) Login(ctx context.Context, req http_auth.LoginRequest) (string, *http_auth.UserResponse, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return "", nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	token, err := s.jwtManager.Generate(user.ID)
	if err != nil {
		return "", nil, fmt.Errorf("generate token: %w", err)
	}

	userResponse := &http_auth.UserResponse{
		ID:          user.ID,
		FullName:    user.FullName,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		IsAdmin:     user.IsAdmin,
	}

	return token, userResponse, nil
}

func (s *AuthService) GenerateToken(ctx context.Context, userID int) (string, error) {
	return s.jwtManager.Generate(userID)
}

func (s *AuthService) AdminExists(ctx context.Context) (bool, error) {
	return s.userRepo.AdminExists(ctx)
}
