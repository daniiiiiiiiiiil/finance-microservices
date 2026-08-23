package service_auth

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string
	Password string
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (string, *UserResponse, error) {
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

	userResponse := &UserResponse{
		ID:          profile.ID,
		FullName:    profile.FullName,
		Email:       profile.Email,
		PhoneNumber: profile.PhoneNumber,
		IsAdmin:     profile.IsAdmin,
	}

	return token, userResponse, nil
}
