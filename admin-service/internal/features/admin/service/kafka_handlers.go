package service_admin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/internal/core/kafka"
	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/internal/core/ports"
	"go.uber.org/zap"
)

// заглушка
func (s *AdminService) StartConsumer(ctx context.Context) {}

func (s *AdminService) HandleTransactionEvent(ctx context.Context, event kafka.Event) error {
	if s.isDuplicate(ctx, event.ID) {
		return nil
	}

	var txEvent kafka.TransactionEvent
	if err := json.Unmarshal(event.Data, &txEvent); err != nil {
		return fmt.Errorf("unmarshal transaction event: %w", err)
	}

	s.logger.Debug("handling transaction event",
		zap.String("type", event.Type),
		zap.Int("transaction_id", txEvent.TransactionID),
		zap.Int("user_id", txEvent.UserID),
	)

	return s.updateMetrics(ctx)
}

func (s *AdminService) HandleTransactionDeleted(ctx context.Context, event kafka.Event) error {
	if s.isDuplicate(ctx, event.ID) {
		return nil
	}

	var txEvent kafka.TransactionEvent
	if err := json.Unmarshal(event.Data, &txEvent); err != nil {
		return fmt.Errorf("unmarshal transaction event: %w", err)
	}

	s.logger.Debug("handling transaction deleted",
		zap.Int("transaction_id", txEvent.TransactionID),
		zap.Int("user_id", txEvent.UserID),
	)

	return s.updateMetrics(ctx)
}

func (s *AdminService) HandleUserCreated(ctx context.Context, event kafka.Event) error {
	if s.isDuplicate(ctx, event.ID) {
		return nil
	}

	var userEvent kafka.UserEvent
	if err := json.Unmarshal(event.Data, &userEvent); err != nil {
		return fmt.Errorf("unmarshal user event: %w", err)
	}

	s.logger.Debug("handling user created",
		zap.Int("user_id", userEvent.UserID),
		zap.String("email", userEvent.Email),
	)

	return s.updateMetrics(ctx)
}

func (s *AdminService) HandleUserDeleted(ctx context.Context, event kafka.Event) error {
	if s.isDuplicate(ctx, event.ID) {
		return nil
	}

	var userEvent kafka.UserEvent
	if err := json.Unmarshal(event.Data, &userEvent); err != nil {
		return fmt.Errorf("unmarshal user event: %w", err)
	}

	s.logger.Debug("handling user deleted",
		zap.Int("user_id", userEvent.UserID),
		zap.String("email", userEvent.Email),
	)

	return s.updateMetrics(ctx)
}

func (s *AdminService) updateMetrics(ctx context.Context) error {
	usersMetrics, err := s.userClient.GetMetrics(ctx)
	if err != nil {
		s.logger.Error("failed to get users metrics", zap.Error(err))
		return fmt.Errorf("get users metrics: %w", err)
	}

	financeMetrics, err := s.financeClient.GetMetrics(ctx)
	if err != nil {
		s.logger.Error("failed to get finance metrics", zap.Error(err))
		return fmt.Errorf("get finance metrics: %w", err)
	}

	metrics := ports.Metrics{
		TotalUsers:        int(usersMetrics.TotalUsers),
		TotalTransactions: financeMetrics.TotalTransactions,
		TotalBalance:      financeMetrics.TotalBalance,
	}

	key := "admin:metrics"
	if err := s.redis.Set(ctx, key, metrics, 10*time.Minute); err != nil {
		s.logger.Error("failed to save metrics to redis", zap.Error(err))
		return fmt.Errorf("set metrics to redis: %w", err)
	}

	s.logger.Debug("metrics updated in redis",
		zap.Int("total_users", metrics.TotalUsers),
		zap.Int("total_transactions", metrics.TotalTransactions),
		zap.Float64("total_balance", metrics.TotalBalance),
	)

	return nil
}

func (s *AdminService) isDuplicate(ctx context.Context, eventID string) bool {
	key := fmt.Sprintf("processed_event:%s", eventID)

	exists, err := s.redis.Exists(ctx, key)
	if err != nil {
		s.logger.Warn("failed to check dedup key", zap.Error(err))
		return false
	}

	if exists > 0 {
		s.logger.Debug("duplicate event ignored", zap.String("event_id", eventID))
		return true
	}

	if err := s.redis.Set(ctx, key, "done", 24*time.Hour); err != nil {
		s.logger.Warn("failed to save dedup key", zap.Error(err))
		return false
	}

	return false
}
