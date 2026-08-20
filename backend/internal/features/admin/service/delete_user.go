package service_admin

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

func (s *AdminService) DeleteUser(ctx context.Context, id int) error {
	s.logger.Info("Starting DeleteUser saga", zap.Int("id", id))

	if err := s.userClient.MarkDeleting(ctx, id); err != nil {
		return fmt.Errorf("marking deleting user failed: %w", err)
	}

	if err := s.financeClient.DeleteUserTransactions(ctx, id); err != nil {
		if restoreErr := s.userClient.RestoreUser(ctx, id); restoreErr != nil {
			s.logger.Error("failed to restore user", zap.Error(restoreErr))
		}
		return fmt.Errorf("marking deleting user failed: %w", err)
	}

	if err := s.userClient.FinalizeDelete(ctx, id); err != nil {
		return fmt.Errorf("finalize delete: %w", err)
	}

	s.logger.Info("saga completed successfully", zap.Int("user_id", id))
	return nil
}
