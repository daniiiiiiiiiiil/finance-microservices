package service_auth

import (
	usersclient "backend/internal/core/clients/users"
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	FullName    string
	Email       string
	Password    string
	PhoneNumber string
	IsAdmin     bool
}

type UserResponse struct {
	ID          int
	FullName    string
	Email       string
	PhoneNumber *string
	IsAdmin     bool
}

func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (string, *UserResponse, error) {
	existing, _ := s.credRepo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return "", nil, fmt.Errorf(`email "%s" already exists`, req.Email)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", nil, fmt.Errorf("failed to hash password: %w", err)
	}

	credID, err := s.credRepo.Create(ctx, req.Email, string(hashedPassword))
	if err != nil {
		return "", nil, fmt.Errorf("failed to create credential: %w", err)
	}

	var phoneNumber *string
	if req.PhoneNumber != "" {
		phoneNumber = &req.PhoneNumber
	}

	profile, err := s.usersClient.CreateProfile(ctx, &usersclient.CreateProfileRequest{
		Email:        req.Email,
		FullName:     req.FullName,
		PhoneNumber:  phoneNumber,
		IsAdmin:      req.IsAdmin,
		PasswordHash: string(hashedPassword),
	})
	if err != nil {
		_ = s.credRepo.AdminUpdateStatus(ctx, credID, "failed")
		return "", nil, fmt.Errorf("failed to create profile: %w", err)
	}
	if err := s.credRepo.AdminUpdateStatus(ctx, credID, "active"); err != nil {
		return "", nil, fmt.Errorf("failed to update credential status: %w", err)
	}

	token, err := s.jwtManager.Generate(profile.ID, profile.Email, profile.IsAdmin)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
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
