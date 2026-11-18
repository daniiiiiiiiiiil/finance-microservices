package postgres

import (
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/domain"
)

type accountModel struct {
	TelegramUserID int64
	ChatID         int64
	UserID         int
	Username       *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func accountDomainFromModel(model accountModel) domain.TelegramAccount {
	return domain.TelegramAccount{
		TelegramUserID: model.TelegramUserID,
		ChatID:         model.ChatID,
		UserID:         model.UserID,
		Username:       model.Username,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}
}

type linkCodeModel struct {
	ID        int64
	UserID    int
	CodeHash  string
	ExpiresAt time.Time
	UsedAt    *time.Time
	Attempts  int
	CreatedAt time.Time
}

func linkCodeDomainFromModel(model linkCodeModel) domain.LinkCode {
	return domain.LinkCode{
		ID:        model.ID,
		UserID:    model.UserID,
		CodeHash:  model.CodeHash,
		ExpiresAt: model.ExpiresAt,
		UsedAt:    model.UsedAt,
		Attempts:  model.Attempts,
		CreatedAt: model.CreatedAt,
	}
}

type notificationSettingsModel struct {
	UserID        int
	BudgetAlerts  bool
	PlanAlerts    bool
	FinanceAlerts bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func notificationSettingsDomainFromModel(model notificationSettingsModel) domain.NotificationSettings {
	return domain.NotificationSettings{
		UserID:        model.UserID,
		BudgetAlerts:  model.BudgetAlerts,
		PlanAlerts:    model.PlanAlerts,
		FinanceAlerts: model.FinanceAlerts,
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
	}
}
