package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/domain"
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
	if err := consumer.RegisterHandler("transaction.created", h.handleTransactionCreated); err != nil {
		return fmt.Errorf("register transaction.created: %w", err)
	}
	if err := consumer.RegisterHandler("transaction.updated", h.handleTransactionUpdated); err != nil {
		return fmt.Errorf("register transaction.updated: %w", err)
	}

	if err := consumer.RegisterHandler("transaction.deleted", h.handleTransactionDeleted); err != nil {
		return fmt.Errorf("register transaction.deleted: %w", err)
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

func (h *FinanceHandlers) handleTransactionCreated(ctx context.Context, event ports.Event) error {
	var tx domain.Finance
	if err := json.Unmarshal(event.Data, &tx); err != nil {
		return fmt.Errorf("parse finance.transaction_created: %w", err)
	}

	h.logger.Info("received finance.transaction_created",
		zap.Int("user_id", tx.UserID),
		zap.Int("transaction_id", tx.ID),
		zap.Float64("amount", tx.Amount))

	if err := h.service.SaveTransactionToAnalytics(ctx, tx); err != nil {
		return fmt.Errorf("save to clickhouse: %w", err)
	}

	return nil
}

func (h *FinanceHandlers) handleTransactionUpdated(ctx context.Context, event ports.Event) error {
	var tx domain.Finance
	if err := json.Unmarshal(event.Data, &tx); err != nil {
		return fmt.Errorf("parse finance.transaction_updated: %w", err)
	}

	h.logger.Info("received finance.transaction_updated",
		zap.Int("user_id", tx.UserID),
		zap.Float64("amount", tx.Amount),
		zap.Int("transaction_id", tx.ID))

	if err := h.service.DeleteTransactionFromAnalytics(ctx, tx.ID); err != nil {
		h.logger.Warn("failed to delete from clickhouse", zap.Error(err))
	}
	if err := h.service.SaveTransactionToAnalytics(ctx, tx); err != nil {
		return fmt.Errorf("save to clickhouse: %w", err)
	}
	return nil
}

func (h *FinanceHandlers) handleTransactionDeleted(ctx context.Context, event ports.Event) error {
	var tx domain.Finance
	if err := json.Unmarshal(event.Data, &tx); err != nil {
		return fmt.Errorf("parse transaction.deleted: %w", err)
	}

	h.logger.Debug("received transaction.deleted",
		zap.Int("transaction_id", tx.ID),
		zap.Int("user_id", tx.UserID))

	if err := h.service.DeleteTransactionFromAnalytics(ctx, tx.ID); err != nil {
		return fmt.Errorf("delete from clickhouse: %w", err)
	}

	return nil
}
