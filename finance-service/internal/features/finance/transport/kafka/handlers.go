package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/ports"
	service_finance "github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/features/finance/service"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/pkg/logger"
	"go.uber.org/zap"
)

type FinanceHandlers struct {
	service *service_finance.FinanceService
	logger  *logger.Logger
}

func NewFinanceHandlers(
	service *service_finance.FinanceService,
	logger *logger.Logger,
) *FinanceHandlers {
	return &FinanceHandlers{
		service: service,
		logger:  logger,
	}
}

func (h *FinanceHandlers) RegisterHandlers(consumer *FinanceKafkaConsumer) error {
	if err := consumer.RegisterHandler("finance.delete_user_transactions", h.handleDeleteUserTransactions); err != nil {
		return fmt.Errorf("register finance.delete_user_transactions: %w", err)
	}
	return nil
}

func (h *FinanceHandlers) handleDeleteUserTransactions(ctx context.Context, event ports.Event) error {
	var req struct {
		UserID int    `json:"user_id"`
		SagaID string `json:"saga_id"`
	}
	if err := json.Unmarshal(event.Data, &req); err != nil {
		return fmt.Errorf("parse finance.delete_user_transactions: %w", err)
	}

	h.logger.Info("received finance.delete_user_transactions",
		zap.Int("user_id", req.UserID),
		zap.String("saga_id", req.SagaID))

	count, err := h.service.DeleteUserTransactions(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("delete user transactions: %w", err)
	}

	h.logger.Info("finance.delete_user_transactions completed",
		zap.Int("user_id", req.UserID),
		zap.Int("deleted_count", count))

	return nil
}
