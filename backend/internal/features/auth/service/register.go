package service_auth

import (
	http_auth "backend/internal/features/auth/transport/http/dto"
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

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
