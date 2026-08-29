package service_user

import (
	"fmt"
	"log"

	"context"
	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/pkg/pagination"
)

func (s *UsersService) ListUsers(ctx context.Context, limit, offset int) ([]domain.User, int, error) {
	limit, offset = pagination.LimitOffset(limit, offset)

	users, found := s.usersListCache.GetUsersList(ctx, limit, offset)
	if found {
		total, err := s.userRepository.GetTotalUsers(ctx)
		if err != nil {
			return nil, 0, fmt.Errorf("get total users: %w", err)
		}
		return users, total, nil
	}

	users, total, err := s.userRepository.ListUsers(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed get users: %w", err)
	}

	go func() {
		if err := s.usersListCache.SetUsersList(context.Background(), users, limit, offset); err != nil {
			log.Printf("failed to cache users list: %v", err)
		}
	}()

	return users, total, nil
}
