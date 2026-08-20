package service_user

import (
	"backend/internal/core/domain"
	"backend/internal/core/kafka"
	"errors"

	"golang.org/x/net/context"
)

func (s *UsersService) CreateProfile(ctx context.Context, req *CreateProfileRequest) (domain.User, error) {
	existing, _ := s.userRepository.GetUserByEmail(ctx, req.Email)
	if existing.ID != 0 {
		return domain.User{}, errors.New("user already exists")
	}

	user := domain.NewUserUninitialized(
		req.FullName,
		req.Email,
		req.PasswordHash,
		req.PhoneNumber,
		req.IsAdmin,
		"active",
	)

	userID, err := s.userRepository.CreateUser(ctx, user)
	if err != nil {
		return domain.User{}, err
	}

	created, err := s.userRepository.GetUser(ctx, userID)
	if err != nil {
		return domain.User{}, err
	}

	go s.sendUserEvent(context.Background(), kafka.EventTypeUserCreated, created.ID, created.Email, created.FullName, created.IsAdmin, created.Status)

	return created, nil
}
