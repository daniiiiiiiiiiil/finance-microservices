package telegram

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/ports"
	apperrors "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/errors"
)

func (r *Router) handleFinanceCallback(ctx context.Context, callback callbackContext, parts []string) error {
	if len(parts) == 1 {
		return r.showFinanceMenu(ctx, callback.ChatID, callback.MessageID)
	}
	switch parts[1] {
	case "dash":
		return r.showFinanceDashboard(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID)
	case "tx":
		return r.handleTransactionCallback(ctx, callback, parts)
	case "bud":
		return r.handleBudgetCallback(ctx, callback, parts)
	case "exp":
		return r.handleExportCallback(ctx, callback, parts)
	case "cats":
		return r.showCategories(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID)
	default:
		return r.showFinanceMenu(ctx, callback.ChatID, callback.MessageID)
	}
}

func (r *Router) showFinanceMenu(ctx context.Context, chatID int64, messageID int) error {
	r.clearDialog(ctx, chatID)
	keyboard := ports.NewKeyboard(
		ports.ButtonRow(
			ports.CallbackButton("📊 Обзор", "fin:dash"),
			ports.CallbackButton("📋 Операции", "fin:tx:all"),
		),
		ports.ButtonRow(
			ports.CallbackButton("➕ Доход / расход", "fin:tx:add"),
			ports.CallbackButton("🎯 Бюджеты", "fin:bud"),
		),
		ports.ButtonRow(
			ports.CallbackButton("📈 Аналитика", "an"),
			ports.CallbackButton("📂 Категории", "fin:cats"),
		),
		ports.ButtonRow(ports.CallbackButton("⬇️ Экспорт", "fin:exp")),
		ports.ButtonRow(ports.CallbackButton("🏠 Главное меню", "menu")),
	)
	return r.editOrSend(ctx, chatID, messageID,
		"<b>Финансы</b>\n\nБаланс, операции, бюджеты, аналитика и экспорт отчётов. Выберите действие:",
		keyboard)
}

func (r *Router) showFinanceDashboard(ctx context.Context, chatID int64, messageID, userID int) error {
	r.clearDialog(ctx, chatID)
	keyboard := ports.NewKeyboard(
		ports.ButtonRow(
			ports.CallbackButton("🔄 Обновить", "fin:dash"),
			ports.CallbackButton("📋 Операции", "fin:tx:all"),
		),
		ports.ButtonRow(
			ports.CallbackButton("🎯 Бюджеты", "fin:bud"),
			ports.CallbackButton("📈 Аналитика", "an"),
		),
		ports.ButtonRow(
			ports.CallbackButton("💰 Финансы", "fin"),
			ports.CallbackButton("🏠 Меню", "menu"),
		),
	)

	if messageID == 0 {
		_ = r.telegram.SendChatAction(ctx, chatID, "typing")
	}

	dashboard, err := r.platform.GetDashboard(ctx, userID)
	if err != nil {
		return r.editOrSend(ctx, chatID, messageID, "Не удалось загрузить финансовый обзор. Попробуйте позже.", keyboard)
	}

	var builder strings.Builder
	builder.WriteString("<b>Финансовый обзор</b>\n\n")
	builder.WriteString(fmt.Sprintf("Баланс: <b>%s ₽</b>\n", formatAmount(dashboard.TotalBalance)))
	builder.WriteString(fmt.Sprintf("Доходы за месяц: %s ₽\n", formatAmount(dashboard.MonthlyIncome)))
	builder.WriteString(fmt.Sprintf("Расходы за месяц: %s ₽\n", formatAmount(dashboard.MonthlyExpenses)))
	builder.WriteString(fmt.Sprintf("Норма сбережений: %.1f%%\n", dashboard.SavingsRate))

	if len(dashboard.RecentTransactions) > 0 {
		builder.WriteString("\n<b>Последние операции</b>\n")
		limit := len(dashboard.RecentTransactions)
		if limit > 5 {
			limit = 5
		}
		for _, transaction := range dashboard.RecentTransactions[:limit] {
			builder.WriteString(fmt.Sprintf(
				"• %s · %s · <b>%s ₽</b> · %s\n",
				formatShortDate(transaction.CreatedAt),
				operationLabel(transaction.Type),
				formatAmount(transaction.Amount),
				escapeHTML(transaction.Category),
			))
		}
	}

	if len(dashboard.TopCategories) > 0 {
		builder.WriteString("\n<b>Топ категорий</b>\n")
		limit := len(dashboard.TopCategories)
		if limit > 5 {
			limit = 5
		}
		for _, category := range dashboard.TopCategories[:limit] {
			builder.WriteString(fmt.Sprintf(
				"• %s — %s ₽ (%.1f%%)\n",
				escapeHTML(category.Category),
				formatAmount(category.Total),
				category.Percentage,
			))
		}
	}

	if len(dashboard.RecentTransactions) == 0 && len(dashboard.TopCategories) == 0 {
		builder.WriteString("\nПока нет операций. Добавьте первый доход или расход.")
	}

	return r.editOrSend(ctx, chatID, messageID, builder.String(), keyboard)
}

func (r *Router) handleAnalyticsCallback(ctx context.Context, callback callbackContext, parts []string) error {
	if len(parts) == 1 {
		return r.showAnalyticsMenu(ctx, callback.ChatID, callback.MessageID)
	}
	period := parts[1]
	switch period {
	case "day", "week", "month", "year":
		return r.showAnalytics(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID, period)
	default:
		return r.showAnalyticsMenu(ctx, callback.ChatID, callback.MessageID)
	}
}

func (r *Router) showAnalyticsMenu(ctx context.Context, chatID int64, messageID int) error {
	r.clearDialog(ctx, chatID)
	keyboard := ports.NewKeyboard(
		ports.ButtonRow(
			ports.CallbackButton("День", "an:day"),
			ports.CallbackButton("Неделя", "an:week"),
		),
		ports.ButtonRow(
			ports.CallbackButton("Месяц", "an:month"),
			ports.CallbackButton("Год", "an:year"),
		),
		ports.ButtonRow(
			ports.CallbackButton("💰 Финансы", "fin"),
			ports.CallbackButton("🏠 Меню", "menu"),
		),
	)
	return r.editOrSend(ctx, chatID, messageID,
		"<b>Аналитика</b>\n\nВыберите период. Данные рассчитываются finance-service по вашим операциям.",
		keyboard)
}

func (r *Router) showAnalytics(ctx context.Context, chatID int64, messageID, userID int, period string) error {
	keyboard := ports.NewKeyboard(
		ports.ButtonRow(
			ports.CallbackButton("День", "an:day"),
			ports.CallbackButton("Неделя", "an:week"),
			ports.CallbackButton("Месяц", "an:month"),
			ports.CallbackButton("Год", "an:year"),
		),
		ports.ButtonRow(
			ports.CallbackButton("🔄 Обновить", "an:"+period),
			ports.CallbackButton("💰 Финансы", "fin"),
		),
		ports.ButtonRow(ports.CallbackButton("🏠 Меню", "menu")),
	)

	analytics, err := r.platform.GetAnalytics(ctx, userID, period)
	if err != nil {
		return r.editOrSend(ctx, chatID, messageID, "Не удалось загрузить аналитику. Попробуйте позже.", keyboard)
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("<b>Аналитика · %s</b>\n\n", analyticsPeriodLabel(period)))
	builder.WriteString(fmt.Sprintf("Доходы: <b>%s ₽</b>\n", formatAmount(analytics.Income)))
	builder.WriteString(fmt.Sprintf("Расходы: <b>%s ₽</b>\n", formatAmount(analytics.Expenses)))
	builder.WriteString(fmt.Sprintf("Итог: <b>%s ₽</b>\n", formatAmount(analytics.Net)))
	builder.WriteString(fmt.Sprintf("Норма сбережений: %.1f%%\n", analytics.SavingsRate))
	if !analytics.From.IsZero() || !analytics.To.IsZero() {
		builder.WriteString(fmt.Sprintf("Период: %s — %s\n", formatShortDate(analytics.From), formatShortDate(analytics.To)))
	}
	if len(analytics.TopCategories) > 0 {
		builder.WriteString("\n<b>Категории</b>\n")
		limit := len(analytics.TopCategories)
		if limit > 7 {
			limit = 7
		}
		for _, category := range analytics.TopCategories[:limit] {
			builder.WriteString(fmt.Sprintf(
				"• %s — %s ₽ (%.1f%%)\n",
				escapeHTML(category.Category),
				formatAmount(category.Total),
				category.Percentage,
			))
		}
	}
	return r.editOrSend(ctx, chatID, messageID, builder.String(), keyboard)
}

func (r *Router) handleExportCallback(ctx context.Context, callback callbackContext, parts []string) error {
	if len(parts) == 2 {
		return r.showExportMenu(ctx, callback.ChatID, callback.MessageID)
	}
	format := parts[2]
	switch format {
	case "csv", "pdf", "json", "txt":
		return r.exportReport(ctx, callback, format)
	default:
		return r.showExportMenu(ctx, callback.ChatID, callback.MessageID)
	}
}

func (r *Router) showExportMenu(ctx context.Context, chatID int64, messageID int) error {
	r.clearDialog(ctx, chatID)
	keyboard := ports.NewKeyboard(
		ports.ButtonRow(
			ports.CallbackButton("CSV", "fin:exp:csv"),
			ports.CallbackButton("PDF", "fin:exp:pdf"),
		),
		ports.ButtonRow(
			ports.CallbackButton("JSON", "fin:exp:json"),
			ports.CallbackButton("TXT", "fin:exp:txt"),
		),
		ports.ButtonRow(
			ports.CallbackButton("💰 Финансы", "fin"),
			ports.CallbackButton("🏠 Меню", "menu"),
		),
	)
	return r.editOrSend(ctx, chatID, messageID,
		"<b>Экспорт отчёта</b>\n\nВыберите формат. Файл будет сформирован finance-service и отправлен в этот чат.",
		keyboard)
}

func (r *Router) exportReport(ctx context.Context, callback callbackContext, format string) error {
	if err := r.telegram.SendChatAction(ctx, callback.ChatID, "upload_document"); err != nil {
		r.logger.Debug("send chat action failed")
	}
	file, err := r.platform.Export(ctx, callback.Account.UserID, format)
	if err != nil {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
			"Не удалось сформировать отчёт. Попробуйте позже.", sectionBack("fin:exp", "Экспорт"))
	}
	download, err := r.platform.DownloadExport(ctx, callback.Account.UserID, file.Key)
	if err != nil {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
			"Отчёт сформирован, но не удалось скачать файл. Попробуйте позже.", sectionBack("fin:exp", "Экспорт"))
	}
	filename := download.Filename
	if filename == "" {
		filename = "finances." + format
	}
	if path.Ext(filename) == "" {
		filename += "." + format
	}
	if len(download.Data) == 0 {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
			"Отчёт пуст: добавьте операции и попробуйте снова.", sectionBack("fin:exp", "Экспорт"))
	}
	if err := r.telegram.SendDocument(ctx, callback.ChatID, download.Data, filename, "Финансовый отчёт · "+strings.ToUpper(format)); err != nil {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
			"Не удалось отправить файл в Telegram.", sectionBack("fin:exp", "Экспорт"))
	}
	return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
		fmt.Sprintf("<b>Отчёт %s отправлен.</b>", strings.ToUpper(format)), sectionBack("fin:exp", "Экспорт"))
}

func (r *Router) showCategories(ctx context.Context, chatID int64, messageID, userID int) error {
	categories, err := r.platform.GetCategories(ctx, userID)
	if err != nil {
		return r.editOrSend(ctx, chatID, messageID, "Не удалось загрузить категории.", sectionBack("fin", "Финансы"))
	}
	var builder strings.Builder
	builder.WriteString("<b>Категории</b>\n\n")
	if len(categories) == 0 {
		builder.WriteString("Пока нет категорий. Создайте операцию — категория появится в списке.")
	} else {
		for _, category := range categories {
			builder.WriteString("• " + escapeHTML(category) + "\n")
		}
	}
	return r.editOrSend(ctx, chatID, messageID, builder.String(), sectionBack("fin", "Финансы"))
}

func (r *Router) handleFinanceFilterText(ctx context.Context, message messageContext, state domain.DialogState) error {
	if state.Step != domain.StepFilterDates {
		return r.showFinanceMenu(ctx, message.ChatID, 0)
	}
	fields := strings.Fields(message.Text)
	if len(fields) != 2 {
		return r.reply(ctx, message.ChatID,
			"Введите период в формате <code>2026-01-01 2026-01-31</code>.",
			ports.NewKeyboard(ports.ButtonRow(ports.CallbackButton("⬅️ Отмена", "fin:tx:all"))))
	}
	from, err := parseDate(fields[0])
	if err != nil {
		return r.reply(ctx, message.ChatID, "Дата начала указана неверно. Формат: 2026-01-01.", nil)
	}
	to, err := parseDate(fields[1])
	if err != nil {
		return r.reply(ctx, message.ChatID, "Дата окончания указана неверно. Формат: 2026-01-31.", nil)
	}
	if to.Before(from) {
		return r.reply(ctx, message.ChatID, "Дата окончания не может быть раньше даты начала.", nil)
	}

	state.Data["from"] = from.Format("2006-01-02")
	state.Data["to"] = to.Format("2006-01-02")
	state.Step = "list"
	if _, err := r.dialogs.Update(ctx, state, "list", state.Data); err != nil {
		return err
	}
	return r.showTransactions(ctx, message.ChatID, 0, message.Account, 0)
}

func analyticsPeriodLabel(period string) string {
	switch period {
	case "day":
		return "день"
	case "week":
		return "неделя"
	case "month":
		return "месяц"
	case "year":
		return "год"
	default:
		return period
	}
}

func mapPlatformError(err error) string {
	switch {
	case errors.Is(err, apperrors.ErrUnauthorized):
		return "Сессия доступа истекла. Попробуйте позже."
	case errors.Is(err, apperrors.ErrForbidden):
		return "Доступ запрещён."
	case errors.Is(err, apperrors.ErrNotFound):
		return "Запись не найдена."
	case errors.Is(err, apperrors.ErrConflict):
		return "Конфликт данных. Обновите экран и попробуйте снова."
	case errors.Is(err, apperrors.ErrInvalidArgument):
		return "Проверьте введённые данные."
	case errors.Is(err, apperrors.ErrUnavailable):
		return "Сервис временно недоступен. Попробуйте позже."
	default:
		return "Не удалось выполнить операцию. Попробуйте позже."
	}
}
