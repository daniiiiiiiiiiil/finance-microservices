package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	corekafka "github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/core/kafka"
	service_auth "github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/features/auth/service"
	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/pkg/logger"
	"go.uber.org/zap"
)

type AuthHandlers struct {
	service *service_auth.AuthService
	logger  *logger.Logger
}

func NewAuthHandlers(
	service *service_auth.AuthService,
	logger *logger.Logger,
) *AuthHandlers {
	return &AuthHandlers{
		service: service,
		logger:  logger,
	}
}

func (h *AuthHandlers) RegisterHandlers(consumer *AuthKafkaConsumer) error {
	if err := consumer.RegisterHandler(corekafka.EventTypeUserDeleted, h.handleUserDeleted); err != nil {
		return fmt.Errorf("register %s: %w", corekafka.EventTypeUserDeleted, err)
	}
	return nil
}

func (h *AuthHandlers) handleUserDeleted(ctx context.Context, event corekafka.Event) error {
	var payload struct {
		UserID int    `json:"user_id"`
		Email  string `json:"email"`
	}
	if err := json.Unmarshal(event.Data, &payload); err != nil {
		return fmt.Errorf("parse user.deleted: %w", err)
	}

	h.logger.Info("received user.deleted",
		zap.Int("user_id", payload.UserID),
		zap.String("email", payload.Email))

	if payload.Email == "" {
		h.logger.Warn("user.deleted has no email, skipping credentials cleanup",
			zap.Int("user_id", payload.UserID))
		return nil
	}

	if err := h.service.DeleteCredentials(ctx, payload.Email); err != nil {
		h.logger.Error("failed to delete credentials",
			zap.Int("user_id", payload.UserID),
			zap.String("email", payload.Email),
			zap.Error(err))
		return fmt.Errorf("delete credentials: %w", err)
	}

	h.logger.Info("credentials deleted successfully",
		zap.Int("user_id", payload.UserID),
		zap.String("email", payload.Email))

	return nil
}
