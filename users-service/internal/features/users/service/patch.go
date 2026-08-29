package service_user

import (
	"fmt"

	"context"
	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/domain"
)

func (s *UsersService) PatchUser(ctx context.Context, id int, patch domain.UserPatch) (domain.User, error) {
	user, err := s.userRepository.GetUser(ctx, id)
	if err != nil {
		return domain.User{}, fmt.Errorf("userRepository.GetUser: %w", err)
	}

	if err := user.ApplyPatch(patch); err != nil {
		return domain.User{}, fmt.Errorf("apply patch: %w", err)
	}

	patchedUser, err := s.userRepository.PatchUser(ctx, id, user)
	if err != nil {
		return domain.User{}, fmt.Errorf("userRepository.PatchUser: %w", err)
	}

	go func() {
		_ = s.userCache.InvalidateUser(context.Background(), id)
		_ = s.usersListCache.InvalidateAllUsersList(context.Background())
	}()

	return patchedUser, nil
}
