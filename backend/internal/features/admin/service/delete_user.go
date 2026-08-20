package service_admin

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

func (s *AdminService) DeleteUser(ctx context.Context, id int, adminID int) error {
	s.logger.Info("Starting DeleteUser saga", zap.Int("id", id))
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	if id == adminID {
		return fmt.Errorf("admin cannot delete themselves")
	}

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
