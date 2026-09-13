package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/ports"
	service_user "github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/features/users/service"
	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/pkg/logger"
	"go.uber.org/zap"
)

type UserHandlers struct {
	service *service_user.UsersService
	logger  *logger.Logger
}

func NewUserHandlers(
	service *service_user.UsersService,
	logger *logger.Logger,
) *UserHandlers {
	return &UserHandlers{
		service: service,
		logger:  logger,
	}
}

func (h *UserHandlers) RegisterHandlers(consumer *UserKafkaConsumer) error {
	if err := consumer.RegisterHandler("user.mark_deleting", h.handleMarkDeleting); err != nil {
		return fmt.Errorf("register user.mark_deleting: %w", err)
	}
	if err := consumer.RegisterHandler("user.restore", h.handleRestoreUser); err != nil {
		return fmt.Errorf("register user.restore: %w", err)
	}
	if err := consumer.RegisterHandler("user.finalize_delete", h.handleFinalizeDelete); err != nil {
		return fmt.Errorf("register user.finalize_delete: %w", err)
	}
	if err := consumer.RegisterHandler("user.create_profile", h.handleCreateProfile); err != nil {
		return fmt.Errorf("register user.create_profile: %w", err)
	}
	if err := consumer.RegisterHandler("user.delete_profile", h.handleDeleteProfile); err != nil {
		return fmt.Errorf("register user.delete_profile: %w", err)
	}
	return nil
}

func (h *UserHandlers) handleMarkDeleting(ctx context.Context, event ports.Event) error {
	var req struct {
		UserID int    `json:"user_id"`
		SagaID string `json:"saga_id"`
	}
	if err := json.Unmarshal(event.Data, &req); err != nil {
		return fmt.Errorf("parse user.mark_deleting: %w", err)
	}

	h.logger.Info("received user.mark_deleting",
		zap.Int("user_id", req.UserID),
		zap.String("saga_id", req.SagaID))

	if err := h.service.MarkDeleting(ctx, req.UserID); err != nil {
		return fmt.Errorf("mark deleting: %w", err)
	}
	return nil
}

func (h *UserHandlers) handleRestoreUser(ctx context.Context, event ports.Event) error {
	var req struct {
		UserID int    `json:"user_id"`
		SagaID string `json:"saga_id"`
	}
	if err := json.Unmarshal(event.Data, &req); err != nil {
		return fmt.Errorf("parse user.restore: %w", err)
	}

	h.logger.Info("received user.restore",
		zap.Int("user_id", req.UserID),
		zap.String("saga_id", req.SagaID))

	if err := h.service.RestoreUser(ctx, req.UserID); err != nil {
		return fmt.Errorf("restore user: %w", err)
	}
	return nil
}

func (h *UserHandlers) handleFinalizeDelete(ctx context.Context, event ports.Event) error {
	var req struct {
		UserID int    `json:"user_id"`
		SagaID string `json:"saga_id"`
	}
	if err := json.Unmarshal(event.Data, &req); err != nil {
		return fmt.Errorf("parse user.finalize_delete: %w", err)
	}

	h.logger.Info("received user.finalize_delete",
		zap.Int("user_id", req.UserID),
		zap.String("saga_id", req.SagaID))

	if err := h.service.FinalizeDelete(ctx, req.UserID); err != nil {
		return fmt.Errorf("finalize delete: %w", err)
	}
	return nil
}

func (h *UserHandlers) handleCreateProfile(ctx context.Context, event ports.Event) error {
	var req struct {
		SagaID       string  `json:"saga_id"`
		Email        string  `json:"email"`
		FullName     string  `json:"full_name"`
		PasswordHash string  `json:"password_hash"`
		PhoneNumber  *string `json:"phone_number,omitempty"`
		IsAdmin      bool    `json:"is_admin"`
	}
	if err := json.Unmarshal(event.Data, &req); err != nil {
		return fmt.Errorf("parse user.create_profile: %w", err)
	}

	h.logger.Info("received user.create_profile",
		zap.String("email", req.Email),
		zap.String("saga_id", req.SagaID))

	_, err := h.service.CreateProfile(ctx, &service_user.CreateProfileRequest{
		Email:        req.Email,
		FullName:     req.FullName,
		PhoneNumber:  req.PhoneNumber,
		IsAdmin:      req.IsAdmin,
		PasswordHash: req.PasswordHash,
	})
	if err != nil {
		return fmt.Errorf("create profile: %w", err)
	}
	return nil
}

func (h *UserHandlers) handleDeleteProfile(ctx context.Context, event ports.Event) error {
	var req struct {
		SagaID string `json:"saga_id"`
		Email  string `json:"email"`
	}
	if err := json.Unmarshal(event.Data, &req); err != nil {
		return fmt.Errorf("parse user.delete_profile: %w", err)
	}

	h.logger.Info("received user.delete_profile",
		zap.String("email", req.Email),
		zap.String("saga_id", req.SagaID))

	user, err := h.service.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return fmt.Errorf("get user by email: %w", err)
	}

	if err := h.service.DeleteUser(ctx, user.ID); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}
