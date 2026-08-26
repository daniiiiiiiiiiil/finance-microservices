package service_user

import (
	"context"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/kafka"
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

	go func() {
		_ = s.userCache.InvalidateUser(context.Background(), id)
		_ = s.usersListCache.InvalidateAllUsersList(context.Background())
	}()

	return nil
}

func (s *UsersService) MarkDeleting(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid user id")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	err = s.userRepository.UpdateStatusTx(ctx, tx, id, "deleting")
	if err != nil {
		return fmt.Errorf("mark deleting: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func (s *UsersService) FinalizeDelete(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid user id")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	user, err := s.userRepository.GetUser(ctx, id)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}

	if user.Status != "deleting" {
		return fmt.Errorf("user status is '%s', expected 'deleting'", user.Status)
	}

	if err := s.userRepository.DeleteUserTx(ctx, tx, id); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	go s.sendUserEvent(context.Background(), kafka.EventTypeUserDeleted, user.ID, user.Email, user.FullName, user.IsAdmin, user.Status)

	return nil
}

func (s *UsersService) RestoreUser(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("id must be positive")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	err = s.userRepository.UpdateStatusTx(ctx, tx, id, "active")
	if err != nil {
		return fmt.Errorf("restore user: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
