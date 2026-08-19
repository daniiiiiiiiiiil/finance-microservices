package service_user

import (
	"backend/internal/core/domain"
	"context"
	"fmt"
	"log"
)

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
