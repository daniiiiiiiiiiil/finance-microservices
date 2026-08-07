package service_admin

import (
	"backend/internal/core/domain"
	"backend/internal/core/pagination"
	"context"
	"fmt"
)

func (s *AdminService) GetUsers(ctx context.Context, limit, offset int) ([]domain.User, int, error) {
	if limit < 0 {
		return nil, 0, fmt.Errorf("limit must be positive")
	}
	if offset < 0 {
		return nil, 0, fmt.Errorf("offset must be positive")
	}
	limit, offset = pagination.LimitOffset(limit, offset)

	users, total, err := s.repo.GetUsers(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("get users: %w", err)
	}
	return users, total, nil
}
