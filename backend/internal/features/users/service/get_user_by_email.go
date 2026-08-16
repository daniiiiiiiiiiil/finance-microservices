package service_user

import (
	"backend/internal/core/domain"
	"context"
	"fmt"
)

func (s *UsersService) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := s.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		return domain.User{}, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}
