package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/ports"
	service "github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/features/service"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/pkg/logger"
	"go.uber.org/zap"
)

type ShoppingHandlers struct {
	service *service.ShoppingService
	logger  *logger.Logger
}

func NewShoppingHandlers(
	service *service.ShoppingService,
	logger *logger.Logger,
) *ShoppingHandlers {
	return &ShoppingHandlers{
		service: service,
		logger:  logger,
	}
}

func (h *ShoppingHandlers) RegisterHandlers(consumer *ShoppingKafkaConsumer) error {
	if err := consumer.RegisterHandler("shopping.delete_user_data", h.handleDeleteUserData); err != nil {
		return fmt.Errorf("register shopping.delete_user_data: %w", err)
	}
	return nil
}

func (h *ShoppingHandlers) handleDeleteUserData(ctx context.Context, event ports.Event) error {
	var req struct {
		UserID int    `json:"user_id"`
		SagaID string `json:"saga_id"`
	}
	if err := json.Unmarshal(event.Data, &req); err != nil {
		return fmt.Errorf("parse shopping.delete_user_data: %w", err)
	}

	h.logger.Info("received shopping.delete_user_data",
		zap.Int("user_id", req.UserID),
		zap.String("saga_id", req.SagaID))

	if err := h.service.DeleteUserData(ctx, req.UserID); err != nil {
		return fmt.Errorf("delete user data: %w", err)
	}

	h.logger.Info("shopping.delete_user_data completed",
		zap.Int("user_id", req.UserID))

	return nil
}
