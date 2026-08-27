package service_user

import (
	"errors"
	"fmt"
	"log"

	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/domain"
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

	go s.publishUserEvent(context.Background(), "user.created", created.ID, created.Email, created.FullName, created.IsAdmin, created.Status)

	return created, nil
}

func (s *UsersService) GetUser(ctx context.Context, id int) (domain.User, error) {
	user, err := s.userCache.GetUser(ctx, id)
	if err == nil {
		return user, nil
	}

	user, err = s.userRepository.GetUser(ctx, id)
	if err != nil {
		return domain.User{}, fmt.Errorf("userRepository.GetUser: %w", err)
	}

	go func() {
		if err := s.userCache.SetUser(context.Background(), user); err != nil {
			log.Printf("failed to cache user %d: %v", id, err)
		}
	}()

	return user, nil
}

func (s *UsersService) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := s.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		return domain.User{}, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}
