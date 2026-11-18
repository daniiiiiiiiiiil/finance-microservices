package service

import (
	"context"
	"errors"
	"fmt"
	"html"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/kafka"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/metrics"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/ports"
	apperrors "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/errors"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/logger"
	"go.uber.org/zap"
)

const (
	NotificationKindBudget  = "budget"
	NotificationKindPlan    = "plan"
	NotificationKindFinance = "finance"
)

type NotificationService struct {
	accounts ports.AccountRepository
	settings ports.NotificationRepository
	telegram ports.TelegramAPI
	logger   *logger.Logger
	retries  int
	wait     func(ctx context.Context, delay time.Duration) error
}

func NewNotificationService(
	accounts ports.AccountRepository,
	settings ports.NotificationRepository,
	telegram ports.TelegramAPI,
	logger *logger.Logger,
	retries int,
) *NotificationService {
	if retries < 1 {
		retries = 1
	}
	return &NotificationService{
		accounts: accounts,
		settings: settings,
		telegram: telegram,
		logger:   logger,
		retries:  retries,
		wait:     waitContext,
	}
}

func (s *NotificationService) Send(ctx context.Context, userID int, notification domain.Notification) error {
	account, err := s.accounts.GetAccountByUserID(ctx, userID)
	if err != nil {
		if isNotFound(err) {
			return nil
		}
		return fmt.Errorf("resolve telegram account: %w", err)
	}

	settings, err := s.settings.GetSettings(ctx, userID)
	if err != nil && !isNotFound(err) {
		return fmt.Errorf("get notification settings: %w", err)
	}
	if isNotFound(err) {
		settings = domain.DefaultNotificationSettings(userID, time.Now().UTC())
	}
	if !notificationEnabled(settings, notification.Kind) {
		return nil
	}

	text := formatNotification(notification)
	var lastErr error
	delay := 500 * time.Millisecond
	for attempt := 1; attempt <= s.retries; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		_, err := s.telegram.SendMessage(ctx, account.ChatID, text, nil)
		if err == nil {
			metrics.RecordTelegramAPI("sendMessage", true)
			metrics.RecordNotification(notification.Kind, true)
			return nil
		}
		lastErr = err
		metrics.RecordTelegramAPI("sendMessage", false)
		if errors.Is(err, apperrors.ErrTelegramForbidden) {
			s.logger.Warn("telegram user blocked the bot", zap.Int("user_id", userID))
			return nil
		}
		if attempt == s.retries {
			break
		}
		if err := s.wait(ctx, delay); err != nil {
			return err
		}
		if delay < 5*time.Second {
			delay *= 2
		}
	}

	metrics.RecordNotification(notification.Kind, false)
	return fmt.Errorf("deliver telegram notification: %w", lastErr)
}

func (s *NotificationService) HandleBudgetExceeded(ctx context.Context, event kafka.BudgetExceededEvent) error {
	if event.UserID <= 0 {
		return fmt.Errorf("user id is required: %w", apperrors.ErrInvalidArgument)
	}
	text := fmt.Sprintf(
		"Категория: <b>%s</b>\nЛимит: %s ₽\nПотрачено: %s ₽\nПревышение: %s ₽",
		html.EscapeString(event.Category),
		formatMoney(event.AmountLimit),
		formatMoney(event.Spent),
		formatMoney(event.ExceededBy),
	)
	if event.Period != "" {
		text += fmt.Sprintf("\nПериод: %s", html.EscapeString(event.Period))
	}
	return s.Send(ctx, event.UserID, domain.Notification{
		Kind:  NotificationKindBudget,
		Title: "Превышен бюджет",
		Text:  text,
	})
}

func (s *NotificationService) HandlePlanCreated(ctx context.Context, event kafka.PlanCreatedEvent) error {
	if event.UserID <= 0 {
		return fmt.Errorf("user id is required: %w", apperrors.ErrInvalidArgument)
	}
	text := fmt.Sprintf(
		"AI-план готов.\nСтоимость: %s ₽\nРекомендация: <b>%s</b>",
		formatMoney(event.TotalCost),
		html.EscapeString(event.Mode),
	)
	return s.Send(ctx, event.UserID, domain.Notification{
		Kind:  NotificationKindPlan,
		Title: "AI-планирование завершено",
		Text:  text,
	})
}

func (s *NotificationService) HandleTransaction(ctx context.Context, event kafka.TransactionEvent) error {
	if event.UserID <= 0 {
		return fmt.Errorf("user id is required: %w", apperrors.ErrInvalidArgument)
	}
	operation := "Расход"
	if event.TypeTransaction == "income" {
		operation = "Доход"
	}
	text := fmt.Sprintf(
		"%s: <b>%s ₽</b>\nКатегория: %s",
		operation,
		formatMoney(event.Amount),
		html.EscapeString(event.Category),
	)
	return s.Send(ctx, event.UserID, domain.Notification{
		Kind:  NotificationKindFinance,
		Title: "Новая операция",
		Text:  text,
	})
}

func (s *NotificationService) HandleUserDeleted(ctx context.Context, event kafka.UserDeletedEvent) error {
	if event.UserID <= 0 {
		return fmt.Errorf("user id is required: %w", apperrors.ErrInvalidArgument)
	}
	if err := s.accounts.DeleteAccountByUserID(ctx, event.UserID); err != nil {
		if !isNotFound(err) {
			return fmt.Errorf("delete telegram account: %w", err)
		}
	}
	if err := s.settings.DeleteSettings(ctx, event.UserID); err != nil {
		s.logger.Warn("failed to delete notification settings", zap.Int("user_id", event.UserID), zap.Error(err))
	}
	s.logger.Info("telegram account removed after user deletion", zap.Int("user_id", event.UserID))
	return nil
}

func notificationEnabled(settings domain.NotificationSettings, kind string) bool {
	switch kind {
	case NotificationKindBudget:
		return settings.BudgetAlerts
	case NotificationKindPlan:
		return settings.PlanAlerts
	case NotificationKindFinance:
		return settings.FinanceAlerts
	default:
		return false
	}
}

func formatNotification(notification domain.Notification) string {
	return fmt.Sprintf("<b>%s</b>\n\n%s", html.EscapeString(notification.Title), notification.Text)
}

func formatMoney(value float64) string {
	return fmt.Sprintf("%.2f", value)
}

func waitContext(ctx context.Context, delay time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
		return nil
	}
}
