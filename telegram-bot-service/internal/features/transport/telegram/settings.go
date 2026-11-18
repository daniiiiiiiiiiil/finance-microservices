package telegram

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/features/service"
)

func (r *Router) handleNotificationsCallback(ctx context.Context, callback callbackContext, parts []string) error {
	if len(parts) == 1 {
		return r.showNotifications(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID)
	}
	if parts[1] == "t" && len(parts) >= 3 {
		return r.toggleNotification(ctx, callback, parts[2])
	}
	return r.showNotifications(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID)
}

func (r *Router) showNotifications(ctx context.Context, chatID int64, messageID, userID int) error {
	r.clearDialog(ctx, chatID)
	settings, err := r.accounts.GetSettings(ctx, userID)
	if err != nil {
		return r.editOrSend(ctx, chatID, messageID, "Не удалось загрузить настройки уведомлений.", sectionBack("set", "Настройки"))
	}
	text := "<b>Уведомления</b>\n\n" +
		"Бот присылает события из платформы через Kafka. Выберите, какие уведомления получать:\n\n" +
		"• <b>Бюджеты</b> — превышение лимита расходов.\n" +
		"• <b>AI-планы</b> — завершение AI-планирования.\n" +
		"• <b>Операции</b> — новые доходы и расходы."
	return r.editOrSend(ctx, chatID, messageID, text,
		ports.NewKeyboard(
			ports.ButtonRow(
				ports.CallbackButton(toggleLabel("Бюджеты", settings.BudgetAlerts), "ntf:t:budget"),
				ports.CallbackButton(toggleLabel("AI-планы", settings.PlanAlerts), "ntf:t:plan"),
			),
			ports.ButtonRow(
				ports.CallbackButton(toggleLabel("Операции", settings.FinanceAlerts), "ntf:t:finance"),
			),
			ports.ButtonRow(
				ports.CallbackButton("⚙️ Настройки", "set"),
				ports.CallbackButton("🏠 Меню", "menu"),
			),
		))
}

func (r *Router) toggleNotification(ctx context.Context, callback callbackContext, kind string) error {
	settings, err := r.accounts.GetSettings(ctx, callback.Account.UserID)
	if err != nil {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, "Не удалось загрузить настройки уведомлений.", sectionBack("set", "Настройки"))
	}
	switch kind {
	case service.NotificationKindBudget:
		settings.BudgetAlerts = !settings.BudgetAlerts
	case service.NotificationKindPlan:
		settings.PlanAlerts = !settings.PlanAlerts
	case service.NotificationKindFinance:
		settings.FinanceAlerts = !settings.FinanceAlerts
	default:
		return r.showNotifications(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID)
	}
	if err := r.accounts.UpdateSettings(ctx, settings); err != nil {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, "Не удалось сохранить настройки.", sectionBack("set", "Настройки"))
	}
	return r.showNotifications(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID)
}

func toggleLabel(name string, enabled bool) string {
	if enabled {
		return "✅ " + name
	}
	return "❌ " + name
}

func (r *Router) handleSettingsCallback(ctx context.Context, callback callbackContext, parts []string) error {
	if len(parts) == 1 {
		return r.showSettings(ctx, callback.ChatID, callback.MessageID, callback.Account)
	}
	switch parts[1] {
	case "unlink":
		if argAt(parts, 2) == "yes" {
			return r.unlinkTelegram(ctx, callback)
		}
		return r.confirmUnlink(ctx, callback.ChatID, callback.MessageID)
	case "ntf":
		return r.showNotifications(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID)
	default:
		return r.showSettings(ctx, callback.ChatID, callback.MessageID, callback.Account)
	}
}

func (r *Router) showSettings(ctx context.Context, chatID int64, messageID int, account domain.TelegramAccount) error {
	r.clearDialog(ctx, chatID)
	var builder strings.Builder
	builder.WriteString("<b>Настройки</b>\n\n")
	profile, err := r.platform.GetProfile(ctx, account.UserID)
	if err == nil {
		builder.WriteString(fmt.Sprintf("Аккаунт: <b>%s</b>\n", escapeHTML(profile.FullName)))
		builder.WriteString(fmt.Sprintf("Email: %s\n", escapeHTML(profile.Email)))
		if profile.IsAdmin {
			builder.WriteString("Роль: администратор\n")
		}
	} else {
		builder.WriteString(fmt.Sprintf("ID аккаунта: %d\n", account.UserID))
	}
	builder.WriteString(fmt.Sprintf("Telegram ID: <code>%d</code>\n", account.TelegramUserID))
	if account.Username != nil {
		builder.WriteString(fmt.Sprintf("Username: @%s\n", escapeHTML(*account.Username)))
	}
	builder.WriteString(fmt.Sprintf("Привязан: %s\n", formatDate(account.CreatedAt)))

	rows := [][]ports.Button{
		ports.ButtonRow(
			ports.CallbackButton("🔔 Уведомления", "ntf"),
			ports.CallbackButton("🔗 Отвязать", "set:unlink"),
		),
	}
	if webApp := r.webAppButton("🌐 Web App"); webApp.Text != "" {
		rows = append(rows, ports.ButtonRow(webApp))
	}
	rows = append(rows, ports.ButtonRow(ports.CallbackButton("🏠 Главное меню", "menu")))
	return r.editOrSend(ctx, chatID, messageID, builder.String(), ports.NewKeyboard(rows...))
}

func (r *Router) unlinkTelegram(ctx context.Context, callback callbackContext) error {
	if err := r.accounts.UnlinkByTelegramID(ctx, callback.TelegramUserID); err != nil {
		if errors.Is(err, service.ErrNotLinked) {
			return r.editOrSend(ctx, callback.ChatID, callback.MessageID, "Telegram уже отвязан.", nil)
		}
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, "Не удалось отвязать Telegram. Попробуйте позже.", nil)
	}
	if err := r.dialogs.Clear(ctx, callback.ChatID); err != nil {
		r.logger.Warn("clear dialog state")
	}
	return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
		"<b>Telegram отвязан.</b>\n\nДоступ к данным через бота закрыт. Привязать аккаунт снова можно через настройки профиля в веб-приложении.",
		nil)
}
