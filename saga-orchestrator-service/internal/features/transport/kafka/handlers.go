package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/features/orchestrator"
	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/pkg/logger"
	"go.uber.org/zap"
)

type SagaHandlers struct {
	deleteUserSaga   *orchestrator.DeleteUserSaga
	registerUserSaga *orchestrator.RegisterUserSaga
	logger           *logger.Logger
}

func NewSagaHandlers(
	deleteUserSaga *orchestrator.DeleteUserSaga,
	registerUserSaga *orchestrator.RegisterUserSaga,
	logger *logger.Logger,
) *SagaHandlers {
	return &SagaHandlers{
		deleteUserSaga:   deleteUserSaga,
		registerUserSaga: registerUserSaga,
		logger:           logger,
	}
}

func (h *SagaHandlers) RegisterHandlers(consumer *SagaKafkaConsumer) error {
	if err := consumer.RegisterHandler(
		EventTypeUserDeleted,
		h.handleUserDeleted,
	); err != nil {
		return fmt.Errorf("register user.deleted handler: %w", err)
	}

	if err := consumer.RegisterHandler(
		EventTypeUserRegisterRequested,
		h.handleUserRegisterRequested,
	); err != nil {
		return fmt.Errorf("register user.register.requested handler: %w", err)
	}

	return nil
}

func (h *SagaHandlers) handleUserDeleted(ctx context.Context, event ports.Event) error {
	userEvent, err := ParseUserDeletedEvent(event.Data)
	if err != nil {
		h.logger.Error("failed to parse user.deleted event",
			zap.String("event_id", event.ID),
			zap.Error(err))
		return fmt.Errorf("parse user.deleted: %w", err)
	}

	h.logger.Info("received user.deleted event",
		zap.Int("user_id", userEvent.UserID),
		zap.String("event_id", event.ID))

	if err := h.deleteUserSaga.HandleUserDeleted(ctx, userEvent.UserID); err != nil {
		h.logger.Error("failed to handle user.deleted",
			zap.Int("user_id", userEvent.UserID),
			zap.Error(err))
		return fmt.Errorf("handle user.deleted: %w", err)
	}

	h.logger.Info("user.deleted processed successfully",
		zap.Int("user_id", userEvent.UserID))

	return nil
}

func (h *SagaHandlers) handleUserRegisterRequested(ctx context.Context, event ports.Event) error {
	h.logger.Info("received user.register.requested event",
		zap.String("event_id", event.ID))

	var req UserRegisterRequestedEvent
	if err := json.Unmarshal(event.Data, &req); err != nil {
		h.logger.Error("failed to parse user.register.requested event",
			zap.String("event_id", event.ID),
			zap.Error(err))
		return fmt.Errorf("parse user.register.requested: %w", err)
	}

	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.FullName == "" {
		return fmt.Errorf("full_name is required")
	}
	if req.PasswordHash == "" {
		return fmt.Errorf("password_hash is required")
	}

	h.logger.Info("starting register user saga",
		zap.String("email", req.Email),
		zap.String("full_name", req.FullName),
		zap.Bool("is_admin", req.IsAdmin))

	if err := h.registerUserSaga.HandleRegisterRequested(
		ctx,
		req.Email,
		req.FullName,
		req.PasswordHash,
		req.PhoneNumber,
		req.IsAdmin,
	); err != nil {
		h.logger.Error("failed to handle user.register.requested",
			zap.String("email", req.Email),
			zap.Error(err))
		return fmt.Errorf("handle user.register.requested: %w", err)
	}

	h.logger.Info("user.register.requested processed successfully",
		zap.String("email", req.Email))

	return nil
}
