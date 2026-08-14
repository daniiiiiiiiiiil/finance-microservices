package service_user

import (
	"backend/internal/core/domain"
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
		"",
		req.PhoneNumber,
		req.IsAdmin,
	)

	userID, err := s.userRepository.CreateUser(ctx, user)
	if err != nil {
		return domain.User{}, err
	}

	created, err := s.userRepository.GetUser(ctx, userID)
	if err != nil {
		return domain.User{}, err
	}

	return created, nil
}
