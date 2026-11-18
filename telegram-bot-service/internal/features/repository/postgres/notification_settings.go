package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/repository/postgres/pool"
	apperrors "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/errors"
)

func (r *Repository) GetSettings(ctx context.Context, userID int) (domain.NotificationSettings, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT user_id, budget_alerts, plan_alerts, finance_alerts, created_at, updated_at
		FROM telegram.notification_settings
		WHERE user_id = $1`

	var model notificationSettingsModel
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&model.UserID,
		&model.BudgetAlerts,
		&model.PlanAlerts,
		&model.FinanceAlerts,
		&model.CreatedAt,
		&model.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pool.ErrNoRows) {
			return domain.NotificationSettings{}, fmt.Errorf("notification settings: %w", apperrors.ErrNotFound)
		}
		return domain.NotificationSettings{}, fmt.Errorf("get notification settings: %w", err)
	}
	return notificationSettingsDomainFromModel(model), nil
}

func (r *Repository) UpsertSettings(ctx context.Context, settings domain.NotificationSettings) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		INSERT INTO telegram.notification_settings (user_id, budget_alerts, plan_alerts, finance_alerts)
		OVERRIDING SYSTEM VALUE
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE SET
			budget_alerts = EXCLUDED.budget_alerts,
			plan_alerts = EXCLUDED.plan_alerts,
			finance_alerts = EXCLUDED.finance_alerts,
			updated_at = NOW()`

	if _, err := r.pool.Exec(ctx, query,
		settings.UserID,
		settings.BudgetAlerts,
		settings.PlanAlerts,
		settings.FinanceAlerts,
	); err != nil {
		return fmt.Errorf("upsert notification settings: %w", err)
	}
	return nil
}

func (r *Repository) DeleteSettings(ctx context.Context, userID int) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	if _, err := r.pool.Exec(ctx, `DELETE FROM telegram.notification_settings WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("delete notification settings: %w", err)
	}
	return nil
}
