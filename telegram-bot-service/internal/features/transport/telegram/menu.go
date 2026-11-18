package telegram

import (
	"context"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/ports"
)

const (
	iconFinance       = "💰"
	iconShopping      = "🛒"
	iconAI            = "🤖"
	iconAnalytics     = "📊"
	iconCurrency      = "💱"
	iconNotifications = "🔔"
	iconSettings      = "⚙️"
	iconBack          = "⬅️"
	iconMenu          = "🏠"
)

func (r *Router) showMenu(ctx context.Context, chatID int64, messageID int) error {
	r.clearDialog(ctx, chatID)
	return r.editOrSend(ctx, chatID, messageID, menuText(), r.mainMenuKeyboard())
}

func (r *Router) mainMenuKeyboard() *ports.Keyboard {
	rows := [][]ports.Button{
		ports.ButtonRow(
			ports.CallbackButton(iconFinance+" Финансы", "fin"),
			ports.CallbackButton(iconShopping+" Списки покупок", "sh"),
		),
		ports.ButtonRow(
			ports.CallbackButton(iconAI+" AI-планировщик", "ai"),
			ports.CallbackButton(iconAnalytics+" Аналитика", "an"),
		),
		ports.ButtonRow(
			ports.CallbackButton(iconCurrency+" Валюты", "cur"),
			ports.CallbackButton(iconNotifications+" Уведомления", "ntf"),
		),
		ports.ButtonRow(
			ports.CallbackButton(iconSettings+" Настройки", "set"),
		),
	}
	if webApp := r.webAppButton("🌐 Web App"); webApp.Text != "" {
		rows = append(rows, ports.ButtonRow(webApp))
	}
	return ports.NewKeyboard(rows...)
}

func (r *Router) webAppButton(text string) ports.Button {
	if !strings.HasPrefix(strings.ToLower(r.webAppURL), "https://") {
		return ports.Button{}
	}
	return ports.WebAppButton(text, r.webAppURL)
}
func backToMenu() *ports.Keyboard {
	return ports.NewKeyboard(
		ports.ButtonRow(ports.CallbackButton(iconMenu+" Главное меню", "menu")),
	)
}

func sectionBack(callback, text string) *ports.Keyboard {
	return ports.NewKeyboard(
		ports.ButtonRow(
			ports.CallbackButton(iconBack+" "+text, callback),
			ports.CallbackButton(iconMenu+" Меню", "menu"),
		),
	)
}

func menuText() string {
	return "<b>FinGoals</b>\n\n" +
		"Финансы, списки покупок, AI-планировщик, аналитика и валюты — всё в одном боте.\n\n" +
		"Выберите раздел:"
}

func helpText() string {
	return "<b>Команды</b>\n\n" +
		"/start — главное меню или привязка аккаунта\n" +
		"/menu — главное меню\n" +
		"/balance — финансовый обзор\n" +
		"/settings — настройки и привязка\n" +
		"/unlink — отвязать Telegram\n" +
		"/help — эта справка\n\n" +
		"Все разделы доступны через кнопки меню. Для привязки аккаунта используйте код из настроек профиля в веб-приложении."
}

func unlinkedText() string {
	return "<b>Аккаунт не привязан</b>\n\n" +
		"1. Войдите в веб-приложение FinGoals.\n" +
		"2. Откройте раздел «Telegram-бот» в боковом меню.\n" +
		"3. Нажмите «Подключить Telegram» и получите код.\n" +
		"4. Отправьте боту команду <code>/start КОД</code>.\n\n" +
		"Пароль в Telegram вводить не нужно."
}

func formatAmount(value float64) string {
	raw := fmt.Sprintf("%.2f", value)
	parts := strings.SplitN(raw, ".", 2)
	integer := parts[0]
	sign := ""
	if strings.HasPrefix(integer, "-") {
		sign = "-"
		integer = strings.TrimPrefix(integer, "-")
	}
	var builder strings.Builder
	builder.WriteString(sign)
	for index, char := range integer {
		if index > 0 && (len(integer)-index)%3 == 0 {
			builder.WriteByte(' ')
		}
		builder.WriteRune(char)
	}
	if len(parts) == 2 {
		builder.WriteByte('.')
		builder.WriteString(parts[1])
	}
	return builder.String()
}

func formatDate(value time.Time) string {
	if value.IsZero() {
		return "—"
	}
	return value.Local().Format("02.01.2006 15:04")
}

func formatShortDate(value time.Time) string {
	if value.IsZero() {
		return "—"
	}
	return value.Local().Format("02.01.2006")
}

func parseDate(value string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", strings.TrimSpace(value), time.Local)
}

func progressBar(current, goal float64) string {
	if goal <= 0 {
		return "▱▱▱▱▱▱▱▱▱▱ 0%"
	}
	ratio := current / goal
	if ratio < 0 {
		ratio = 0
	}
	const segments = 10
	filled := int(ratio * segments)
	if filled > segments {
		filled = segments
	}
	if filled < 0 {
		filled = 0
	}
	return strings.Repeat("▰", filled) + strings.Repeat("▱", segments-filled) + fmt.Sprintf(" %.0f%%", ratio*100)
}

func escapeHTML(value string) string {
	return html.EscapeString(value)
}

func truncate(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	if limit <= 1 {
		return string(runes[:limit])
	}
	return string(runes[:limit-1]) + "…"
}

func operationLabel(transactionType string) string {
	if transactionType == domain.ActionIncome {
		return "Доход"
	}
	return "Расход"
}
