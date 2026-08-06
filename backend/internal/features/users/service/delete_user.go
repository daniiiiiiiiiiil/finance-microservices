package service

import (
	"context"
	"fmt"
)

func (s *UsersService) DeleteUser(ctx context.Context, id int) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	if err := s.userRepository.DeleteUserTx(ctx, tx, id); err != nil {
		return fmt.Errorf("DeleteUser: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
