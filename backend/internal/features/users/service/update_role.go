package service_user

import (
	"backend/internal/core/domain"
	"fmt"

	"golang.org/x/net/context"
)

func (s *UsersService) UpdateRoleUsers(ctx context.Context, id int, isAdmin bool) (domain.User, error) {
	if id <= 0 {
		return domain.User{}, fmt.Errorf("id must be positive")
	}
	user, err := s.GetUser(ctx, id)
	if err != nil {
		return domain.User{}, fmt.Errorf("Failed get user: %w", err)
	}
	if err := user.ApplyPatchRole(isAdmin); err != nil {
		return domain.User{}, fmt.Errorf("apply patch role: %w", err)
	}

	update, err := s.userRepository.UpdateRoleUsers(ctx, id, isAdmin)
	if err != nil {
		return domain.User{}, fmt.Errorf("Failed update user role err: %w", err)
	}
	return update, nil
}
