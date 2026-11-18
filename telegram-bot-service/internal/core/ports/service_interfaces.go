package ports

import (
	"context"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/kafka"
)

type LinkCodeResult struct {
	Code      string
	ExpiresAt time.Time
}

type AccountServiceInterface interface {
	CreateLinkCode(ctx context.Context, userID int) (LinkCodeResult, error)
	ConsumeLinkCode(ctx context.Context, code string, telegramUserID, chatID int64, username *string) (domain.TelegramAccount, error)
	GetLink(ctx context.Context, userID int) (domain.TelegramAccount, error)
	Resolve(ctx context.Context, telegramUserID int64) (domain.TelegramAccount, error)
	UnlinkByTelegramID(ctx context.Context, telegramUserID int64) error
	UnlinkByUserID(ctx context.Context, userID int) error
	GetSettings(ctx context.Context, userID int) (domain.NotificationSettings, error)
	UpdateSettings(ctx context.Context, settings domain.NotificationSettings) error
}

type DialogServiceInterface interface {
	Get(ctx context.Context, chatID int64) (domain.DialogState, bool)
	Start(ctx context.Context, chatID int64, userID int, flow domain.Flow, step string, data map[string]string) (domain.DialogState, error)
	Update(ctx context.Context, state domain.DialogState, step string, data map[string]string) (domain.DialogState, error)
	Clear(ctx context.Context, chatID int64) error
}

type NotificationServiceInterface interface {
	Send(ctx context.Context, userID int, notification domain.Notification) error
	HandleBudgetExceeded(ctx context.Context, event kafka.BudgetExceededEvent) error
	HandlePlanCreated(ctx context.Context, event kafka.PlanCreatedEvent) error
	HandleTransaction(ctx context.Context, event kafka.TransactionEvent) error
	HandleUserDeleted(ctx context.Context, event kafka.UserDeletedEvent) error
}
