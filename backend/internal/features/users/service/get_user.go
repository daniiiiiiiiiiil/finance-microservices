package service_user

import (
	"backend/internal/core/domain"
	"context"
	"fmt"
)

func (s *UsersService) GetUser(ctx context.Context, id int) (domain.User, error) {
	user, err := s.userRepository.GetUser(ctx, id)
	if err != nil {
		return domain.User{}, fmt.Errorf("userRepository.GetUser: %w", err)
	}
	return user, nil
}
