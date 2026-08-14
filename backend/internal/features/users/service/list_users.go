package service_user

import (
	"backend/internal/core/domain"
	"backend/internal/core/pagination"
	"fmt"

	"golang.org/x/net/context"
)

func (s *UsersService) ListUsers(ctx context.Context, limit, offset int) ([]domain.User, int, error) {
	limit, offset = pagination.LimitOffset(limit, offset)
	list, total, err := s.userRepository.ListUsers(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("Failed get users err: %w", err)
	}
	return list, total, nil
}
