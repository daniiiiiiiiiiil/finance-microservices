package domain

import "time"

type TelegramAccount struct {
	TelegramUserID int64
	ChatID         int64
	UserID         int
	Username       *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type LinkCode struct {
	ID        int64
	UserID    int
	CodeHash  string
	ExpiresAt time.Time
	UsedAt    *time.Time
	Attempts  int
	CreatedAt time.Time
}

func (c LinkCode) Expired(now time.Time) bool {
	return !now.Before(c.ExpiresAt)
}

func (c LinkCode) Used() bool {
	return c.UsedAt != nil
}

type NotificationSettings struct {
	UserID        int
	BudgetAlerts  bool
	PlanAlerts    bool
	FinanceAlerts bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func DefaultNotificationSettings(userID int, now time.Time) NotificationSettings {
	return NotificationSettings{
		UserID:        userID,
		BudgetAlerts:  true,
		PlanAlerts:    true,
		FinanceAlerts: false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}
