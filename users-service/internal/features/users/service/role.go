package service_user

import (
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/domain"
	"golang.org/x/net/context"
)

func (s *UsersService) UpdateRole(ctx context.Context, id int, isAdmin bool) (domain.User, error) {
	if id <= 0 {
		return domain.User{}, fmt.Errorf("id must be positive")
	}

	user, err := s.GetUser(ctx, id)
	if err != nil {
		return domain.User{}, fmt.Errorf("failed get user: %w", err)
	}

	if err := user.ApplyPatchRole(isAdmin); err != nil {
		return domain.User{}, fmt.Errorf("apply patch role: %w", err)
	}

	update, err := s.userRepository.UpdateRole(ctx, id, isAdmin)
	if err != nil {
		return domain.User{}, fmt.Errorf("failed update user role: %w", err)
	}

	go func() {
		_ = s.userCache.InvalidateUser(context.Background(), id)
		_ = s.usersListCache.InvalidateAllUsersList(context.Background())
	}()

	return update, nil
}
