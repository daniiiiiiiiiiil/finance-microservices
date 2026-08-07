package service_user

import (
	"backend/internal/core/domain"
	"context"
	"fmt"
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

	return patchedUser, nil
}
