package service_auth

import (
	"context"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/core/ports"
	"golang.org/x/crypto/bcrypt"
)

func (s *AuthService) Login(ctx context.Context, req ports.LoginRequest) (string, *ports.UserResponse, error) {
	cred, err := s.credRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return "", nil, ErrInvalidCredentials
	}

	if cred.Status != "active" {
		return "", nil, fmt.Errorf("account not active")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(cred.PasswordHash), []byte(req.Password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	profile, err := s.usersClient.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	token, err := s.jwtManager.Generate(profile.ID, profile.Email, profile.IsAdmin)
	if err != nil {
		return "", nil, fmt.Errorf("generate token: %w", err)
	}

	userResponse := &ports.UserResponse{
		ID:          profile.ID,
		FullName:    profile.FullName,
		Email:       profile.Email,
		PhoneNumber: profile.PhoneNumber,
		IsAdmin:     profile.IsAdmin,
	}

	return token, userResponse, nil
}
