package service_auth

import (
	http_auth "backend/internal/features/auth/transport/http/dto"
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

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
