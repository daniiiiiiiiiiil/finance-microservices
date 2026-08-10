package service_admin

import (
	"backend/internal/core/domain"
	"context"
	"fmt"
)

func (s *AdminService) UpdateUserRole(ctx context.Context, id int, isAdmin bool) (domain.User, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return domain.User{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	user, err := s.repo.UpdateUserRoleTx(ctx, tx, id, isAdmin)
	if err != nil {
		return domain.User{}, fmt.Errorf("update user role: %w", err)
	}

	go s.invalidateCache(context.Background(), id)
	return user, nil
}
