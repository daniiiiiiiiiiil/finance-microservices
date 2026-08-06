package service

import (
	"backend/internal/core/domain"
	"context"
	"fmt"
)

func (s *AdminService) GetUser(ctx context.Context, id int) (domain.User, error) {
	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		return domain.User{}, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}
