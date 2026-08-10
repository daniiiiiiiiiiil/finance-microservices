package service_admin

import (
	"context"
	"fmt"
)

func (s *AdminService) DeleteUser(ctx context.Context, id int) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := s.repo.DeleteUserTx(ctx, tx, id); err != nil {
		return fmt.Errorf("DeleteUserTx: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("Commit: %w", err)
	}
	go s.invalidateCache(context.Background(), id)

	return nil
}
