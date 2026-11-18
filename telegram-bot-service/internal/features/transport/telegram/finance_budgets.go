package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/ports"
)

func (r *Router) handleBudgetCallback(ctx context.Context, callback callbackContext, parts []string) error {
	if len(parts) < 3 {
		return r.showBudgets(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID)
	}
	switch parts[2] {
	case "create":
		return r.startBudgetCategory(ctx, callback, 0)
	case "cat":
		return r.selectBudgetCategory(ctx, callback, argAt(parts, 3))
	case "period":
		return r.selectBudgetPeriod(ctx, callback, argAt(parts, 3))
	case "confirm":
		return r.confirmBudget(ctx, callback)
	case "cancel":
		if err := r.dialogs.Clear(ctx, callback.ChatID); err != nil {
			r.logger.Warn("clear dialog state")
		}
		return r.showBudgets(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID)
	case "e":
		return r.startBudgetEdit(ctx, callback, parseID(argAt(parts, 3)))
	case "d":
		return r.handleDeleteBudget(ctx, callback, parts)
	default:
		return r.showBudgets(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID)
	}
}

func (r *Router) showBudgets(ctx context.Context, chatID int64, messageID, userID int) error {
	budgets, err := r.platform.ListBudgets(ctx, userID)
	if err != nil {
		return r.editOrSend(ctx, chatID, messageID, "Не удалось загрузить бюджеты. Попробуйте позже.", sectionBack("fin", "Финансы"))
	}

	var builder strings.Builder
	builder.WriteString("<b>Бюджеты</b>\n\n")
	rows := make([][]ports.Button, 0, len(budgets)+4)
	if len(budgets) == 0 {
		builder.WriteString("Бюджеты не заданы. Создайте лимит расходов по категории.")
	} else {
		for _, status := range budgets {
			builder.WriteString(fmt.Sprintf(
				"<b>%s</b> · %s\n%s\n%s / %s ₽",
				escapeHTML(status.Budget.Category),
				budgetPeriodLabel(status.Budget.Period),
				progressBar(status.Spent, status.Budget.AmountLimit),
				formatAmount(status.Spent),
				formatAmount(status.Budget.AmountLimit),
			))
			if status.Exceeded {
				builder.WriteString(fmt.Sprintf("\n⚠️ Превышен на %s ₽", formatAmount(-status.Remaining)))
			} else {
				builder.WriteString(fmt.Sprintf("\nОсталось: %s ₽", formatAmount(status.Remaining)))
			}
			builder.WriteString("\n\n")
			rows = append(rows, ports.ButtonRow(
				ports.CallbackButton("✏️ "+truncate(status.Budget.Category, 18), fmt.Sprintf("fin:bud:e:%d", status.Budget.ID)),
				ports.CallbackButton("🗑 Удалить", fmt.Sprintf("fin:bud:d:%d", status.Budget.ID)),
			))
		}
	}
	rows = append(rows, ports.ButtonRow(ports.CallbackButton("➕ Создать бюджет", "fin:bud:create")))
	rows = append(rows, ports.ButtonRow(
		ports.CallbackButton("💰 Финансы", "fin"),
		ports.CallbackButton("🏠 Меню", "menu"),
	))
	return r.editOrSend(ctx, chatID, messageID, builder.String(), ports.NewKeyboard(rows...))
}

func (r *Router) startBudgetCategory(ctx context.Context, callback callbackContext, budgetID int) error {
	data := map[string]string{}
	if budgetID > 0 {
		data["budget_id"] = strconv.Itoa(budgetID)
	}
	if _, err := r.dialogs.Start(ctx, callback.ChatID, callback.Account.UserID, domain.FlowBudget, domain.StepBudgetCategory, data); err != nil {
		return err
	}
	keyboard := r.categoryKeyboard(ctx, callback.Account.UserID, "fin:bud:cat", true)
	return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
		"<b>Бюджет</b>\n\nВыберите категорию или отправьте своё название сообщением:", keyboard)
}

func (r *Router) selectBudgetCategory(ctx context.Context, callback callbackContext, value string) error {
	state, active := r.dialogs.Get(ctx, callback.ChatID)
	if !active || state.Flow != domain.FlowBudget || state.UserID != callback.Account.UserID {
		return r.showBudgets(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID)
	}
	if state.Step != domain.StepBudgetCategory {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, "Категория уже выбрана.", nil)
	}
	if value == "custom" {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, "Отправьте название категории сообщением.", nil)
	}
	categories := r.categories(ctx, callback.Account.UserID)
	index, err := strconv.Atoi(value)
	if err != nil || index < 0 || index >= len(categories) {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, "Категория не найдена.", nil)
	}
	state.Data["category"] = categories[index]
	state.Step = domain.StepBudgetAmount
	if _, err := r.dialogs.Update(ctx, state, domain.StepBudgetAmount, state.Data); err != nil {
		return err
	}
	return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
		fmt.Sprintf("<b>Категория:</b> %s\n\nВведите лимит расходов в рублях, например <code>10000</code>:", escapeHTML(categories[index])),
		ports.NewKeyboard(ports.ButtonRow(ports.CallbackButton("❌ Отмена", "fin:bud:cancel"))))
}

func (r *Router) startBudgetEdit(ctx context.Context, callback callbackContext, budgetID int) error {
	if budgetID <= 0 {
		return r.showBudgets(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID)
	}
	budgets, err := r.platform.ListBudgets(ctx, callback.Account.UserID)
	if err != nil {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, "Не удалось загрузить бюджет.", sectionBack("fin:bud", "Бюджеты"))
	}
	var selected *domain.BudgetStatus
	for index := range budgets {
		if budgets[index].Budget.ID == budgetID {
			selected = &budgets[index]
			break
		}
	}
	if selected == nil {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, "Бюджет не найден.", sectionBack("fin:bud", "Бюджеты"))
	}
	data := map[string]string{
		"budget_id": strconv.Itoa(selected.Budget.ID),
		"category":  selected.Budget.Category,
		"period":    selected.Budget.Period,
	}
	if _, err := r.dialogs.Start(ctx, callback.ChatID, callback.Account.UserID, domain.FlowBudget, domain.StepBudgetAmount, data); err != nil {
		return err
	}
	return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
		fmt.Sprintf("<b>%s</b>\n\nТекущий лимит: %s ₽.\nВведите новый лимит:", escapeHTML(selected.Budget.Category), formatAmount(selected.Budget.AmountLimit)),
		ports.NewKeyboard(ports.ButtonRow(ports.CallbackButton("❌ Отмена", "fin:bud"))))
}

func (r *Router) handleBudgetText(ctx context.Context, message messageContext, state domain.DialogState) error {
	switch state.Step {
	case domain.StepBudgetCategory:
		category := strings.TrimSpace(message.Text)
		if len([]rune(category)) < 1 || len([]rune(category)) > 60 {
			return r.reply(ctx, message.ChatID, "Название категории должно быть от 1 до 60 символов.", nil)
		}
		state.Data["category"] = category
		state.Step = domain.StepBudgetAmount
		if _, err := r.dialogs.Update(ctx, state, domain.StepBudgetAmount, state.Data); err != nil {
			return err
		}
		return r.reply(ctx, message.ChatID,
			fmt.Sprintf("<b>Категория:</b> %s\n\nВведите лимит расходов в рублях:", escapeHTML(category)),
			ports.NewKeyboard(ports.ButtonRow(ports.CallbackButton("❌ Отмена", "fin:bud:cancel"))))
	case domain.StepBudgetAmount:
		amount, err := parseAmount(message.Text)
		if err != nil {
			return r.reply(ctx, message.ChatID, "Лимит должен быть положительным числом, например <code>10000</code>.", nil)
		}
		state.Data["amount"] = strconv.FormatFloat(amount, 'f', 2, 64)
		state.Step = domain.StepBudgetPeriod
		if _, err := r.dialogs.Update(ctx, state, domain.StepBudgetPeriod, state.Data); err != nil {
			return err
		}
		return r.reply(ctx, message.ChatID, "Выберите период бюджета:",
			ports.NewKeyboard(
				ports.ButtonRow(
					ports.CallbackButton("День", "fin:bud:period:day"),
					ports.CallbackButton("Неделя", "fin:bud:period:week"),
				),
				ports.ButtonRow(
					ports.CallbackButton("Месяц", "fin:bud:period:month"),
					ports.CallbackButton("Год", "fin:bud:period:year"),
				),
				ports.ButtonRow(ports.CallbackButton("❌ Отмена", "fin:bud:cancel")),
			))
	default:
		return r.reply(ctx, message.ChatID, "Используйте кнопки в сообщении выше.", nil)
	}
}

func (r *Router) selectBudgetPeriod(ctx context.Context, callback callbackContext, period string) error {
	state, active := r.dialogs.Get(ctx, callback.ChatID)
	if !active || state.Flow != domain.FlowBudget || state.UserID != callback.Account.UserID {
		return r.showBudgets(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID)
	}
	switch period {
	case "day", "week", "month", "year":
	default:
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, "Неверный период.", nil)
	}
	state.Data["period"] = period
	state.Step = domain.StepBudgetConfirm
	if _, err := r.dialogs.Update(ctx, state, domain.StepBudgetConfirm, state.Data); err != nil {
		return err
	}
	action := "Создать"
	if state.Get("budget_id") != "" {
		action = "Обновить"
	}
	text := fmt.Sprintf(
		"<b>Проверьте бюджет</b>\n\nКатегория: %s\nЛимит: <b>%s ₽</b>\nПериод: %s",
		escapeHTML(state.Get("category")),
		formatAmount(parseAmountMust(state.Get("amount"))),
		budgetPeriodLabel(period),
	)
	return r.editOrSend(ctx, callback.ChatID, callback.MessageID, text,
		ports.NewKeyboard(
			ports.ButtonRow(
				ports.CallbackButton("✅ "+action, "fin:bud:confirm"),
				ports.CallbackButton("❌ Отмена", "fin:bud:cancel"),
			),
		))
}

func (r *Router) confirmBudget(ctx context.Context, callback callbackContext) error {
	state, active := r.dialogs.Get(ctx, callback.ChatID)
	if !active || state.Flow != domain.FlowBudget || state.UserID != callback.Account.UserID || state.Step != domain.StepBudgetConfirm {
		return r.showBudgets(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID)
	}
	category := state.Get("category")
	amount := parseAmountMust(state.Get("amount"))
	period := state.Get("period")
	if category == "" || amount <= 0 || period == "" {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, "Черновик бюджета повреждён. Начните заново.", sectionBack("fin:bud", "Бюджеты"))
	}

	if err := r.dialogs.Clear(ctx, callback.ChatID); err != nil {
		r.logger.Warn("clear dialog state")
	}

	budgetID := parseID(state.Get("budget_id"))
	var err error
	if budgetID > 0 {
		_, err = r.platform.UpdateBudget(ctx, callback.Account.UserID, budgetID, category, amount, period)
	} else {
		_, err = r.platform.CreateBudget(ctx, callback.Account.UserID, category, amount, period)
	}
	if err != nil {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
			"Не удалось сохранить бюджет. "+mapPlatformError(err), sectionBack("fin:bud", "Бюджеты"))
	}
	action := "создан"
	if budgetID > 0 {
		action = "обновлён"
	}
	return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
		fmt.Sprintf("<b>Бюджет %s.</b>\n\n%s · %s ₽ · %s", action, escapeHTML(category), formatAmount(amount), budgetPeriodLabel(period)),
		ports.NewKeyboard(
			ports.ButtonRow(
				ports.CallbackButton("🎯 К бюджетам", "fin:bud"),
				ports.CallbackButton("🏠 Меню", "menu"),
			),
		))
}

func (r *Router) handleDeleteBudget(ctx context.Context, callback callbackContext, parts []string) error {
	id := parseID(argAt(parts, 3))
	if id <= 0 {
		return r.showBudgets(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID)
	}
	if argAt(parts, 4) != "yes" {
		budgets, err := r.platform.ListBudgets(ctx, callback.Account.UserID)
		if err != nil {
			return r.editOrSend(ctx, callback.ChatID, callback.MessageID, "Не удалось загрузить бюджет.", sectionBack("fin:bud", "Бюджеты"))
		}
		var selected *domain.BudgetStatus
		for index := range budgets {
			if budgets[index].Budget.ID == id {
				selected = &budgets[index]
				break
			}
		}
		if selected == nil {
			return r.editOrSend(ctx, callback.ChatID, callback.MessageID, "Бюджет не найден.", sectionBack("fin:bud", "Бюджеты"))
		}
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
			fmt.Sprintf("<b>Удалить бюджет?</b>\n\n%s · %s ₽ · %s", escapeHTML(selected.Budget.Category), formatAmount(selected.Budget.AmountLimit), budgetPeriodLabel(selected.Budget.Period)),
			ports.NewKeyboard(
				ports.ButtonRow(
					ports.CallbackButton("🗑 Да, удалить", fmt.Sprintf("fin:bud:d:%d:yes", id)),
					ports.CallbackButton("⬅️ Отмена", "fin:bud"),
				),
			))
	}
	if err := r.platform.DeleteBudget(ctx, callback.Account.UserID, id); err != nil {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, mapPlatformError(err), sectionBack("fin:bud", "Бюджеты"))
	}
	return r.editOrSend(ctx, callback.ChatID, callback.MessageID, "<b>Бюджет удалён.</b>", sectionBack("fin:bud", "Бюджеты"))
}

func budgetPeriodLabel(period string) string {
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
