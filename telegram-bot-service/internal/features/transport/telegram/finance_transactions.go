package telegram

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/pagination"
)

var defaultCategories = []string{
	"Продукты",
	"Транспорт",
	"Жильё",
	"Развлечения",
	"Здоровье",
	"Образование",
	"Одежда",
	"Кафе",
	"Связь",
	"Другое",
}

func (r *Router) handleTransactionCallback(ctx context.Context, callback callbackContext, parts []string) error {
	if len(parts) < 3 {
		return r.showFinanceMenu(ctx, callback.ChatID, callback.MessageID)
	}
	switch parts[2] {
	case "all":
		if err := r.saveFilter(ctx, callback, domain.TransactionFilter{Limit: 10}); err != nil {
			return err
		}
		return r.showTransactions(ctx, callback.ChatID, callback.MessageID, callback.Account, 0)
	case "page":
		offset := parseID(argAt(parts, 3))
		return r.showTransactions(ctx, callback.ChatID, callback.MessageID, callback.Account, offset)
	case "add":
		return r.showTransactionTypeSelection(ctx, callback.ChatID, callback.MessageID)
	case "new":
		return r.startNewTransaction(ctx, callback, argAt(parts, 3))
	case "cat":
		return r.selectTransactionCategory(ctx, callback, argAt(parts, 3))
	case "skipnote":
		return r.confirmTransactionDraft(ctx, callback, "")
	case "confirm":
		return r.confirmTransaction(ctx, callback)
	case "cancel":
		if err := r.dialogs.Clear(ctx, callback.ChatID); err != nil {
			r.logger.Warn("clear dialog state")
		}
		return r.showFinanceMenu(ctx, callback.ChatID, callback.MessageID)
	case "view":
		return r.showTransaction(ctx, callback, parseID(argAt(parts, 3)))
	case "edit":
		return r.startEditTransaction(ctx, callback, parseID(argAt(parts, 3)))
	case "del":
		return r.handleDeleteTransaction(ctx, callback, parts)
	case "filter":
		return r.showFilterMenu(ctx, callback.ChatID, callback.MessageID)
	case "fdate":
		return r.startDateFilter(ctx, callback)
	case "fcat":
		return r.showFilterCategories(ctx, callback)
	case "fc":
		return r.applyFilterCategory(ctx, callback, argAt(parts, 3))
	case "ft":
		return r.applyFilterType(ctx, callback, argAt(parts, 3))
	default:
		return r.showFinanceMenu(ctx, callback.ChatID, callback.MessageID)
	}
}

func (r *Router) showTransactions(ctx context.Context, chatID int64, messageID int, account domain.TelegramAccount, offset int) error {
	filter := r.currentFilter(ctx, chatID, account.UserID)
	limit := 10
	if filter.Limit > 0 {
		limit = filter.Limit
	}
	offset = normalizeOffset(offset)
	filter.Limit = limit
	filter.Offset = offset

	page, err := r.platform.GetTransactions(ctx, account.UserID, filter)
	if err != nil {
		return r.editOrSend(ctx, chatID, messageID, "Не удалось загрузить операции. Попробуйте позже.", sectionBack("fin", "Финансы"))
	}
	if len(page.Items) == 0 {
		return r.editOrSend(ctx, chatID, messageID,
			"<b>Операции</b>\n\nПо выбранным условиям операций нет.",
			ports.NewKeyboard(
				ports.ButtonRow(
					ports.CallbackButton("➕ Добавить", "fin:tx:add"),
					ports.CallbackButton("🔎 Фильтры", "fin:tx:filter"),
				),
				ports.ButtonRow(
					ports.CallbackButton("💰 Финансы", "fin"),
					ports.CallbackButton("🏠 Меню", "menu"),
				),
			))
	}

	meta := pagination.NewPagination(page.Total, limit, offset)
	var builder strings.Builder
	builder.WriteString("<b>Операции</b>\n")
	builder.WriteString(fmt.Sprintf("Страница %d из %d · всего %d\n", meta.CurrentPage, maxInt(meta.TotalPages, 1), page.Total))
	builder.WriteString(filterDescription(filter))
	builder.WriteString("\n")

	rows := make([][]ports.Button, 0, len(page.Items)+4)
	for _, transaction := range page.Items {
		builder.WriteString(fmt.Sprintf(
			"<b>#%d</b> %s · %s · %s ₽ · %s\n",
			transaction.ID,
			formatShortDate(transaction.CreatedAt),
			operationLabel(transaction.Type),
			formatAmount(transaction.Amount),
			escapeHTML(transaction.Category),
		))
		rows = append(rows, ports.ButtonRow(
			ports.CallbackButton(fmt.Sprintf("#%d · %s %s ₽", transaction.ID, operationLabel(transaction.Type), formatAmount(transaction.Amount)), fmt.Sprintf("fin:tx:view:%d", transaction.ID)),
		))
	}

	navRow := make([]ports.Button, 0, 2)
	if meta.HasPrev {
		navRow = append(navRow, ports.CallbackButton("⬅️ Назад", fmt.Sprintf("fin:tx:page:%d", offset-limit)))
	}
	if meta.HasNext {
		navRow = append(navRow, ports.CallbackButton("Вперёд ➡️", fmt.Sprintf("fin:tx:page:%d", offset+limit)))
	}
	if len(navRow) > 0 {
		rows = append(rows, navRow)
	}
	rows = append(rows, ports.ButtonRow(
		ports.CallbackButton("➕ Добавить", "fin:tx:add"),
		ports.CallbackButton("🔎 Фильтры", "fin:tx:filter"),
	))
	rows = append(rows, ports.ButtonRow(
		ports.CallbackButton("💰 Финансы", "fin"),
		ports.CallbackButton("🏠 Меню", "menu"),
	))

	return r.editOrSend(ctx, chatID, messageID, builder.String(), ports.NewKeyboard(rows...))
}

func (r *Router) showTransactionTypeSelection(ctx context.Context, chatID int64, messageID int) error {
	return r.editOrSend(ctx, chatID, messageID,
		"<b>Новая операция</b>\n\nЧто вы хотите добавить?",
		ports.NewKeyboard(
			ports.ButtonRow(
				ports.CallbackButton("📈 Доход", "fin:tx:new:income"),
				ports.CallbackButton("📉 Расход", "fin:tx:new:expense"),
			),
			ports.ButtonRow(
				ports.CallbackButton("⬅️ Назад", "fin:tx:all"),
				ports.CallbackButton("🏠 Меню", "menu"),
			),
		))
}

func (r *Router) startNewTransaction(ctx context.Context, callback callbackContext, transactionType string) error {
	if transactionType != domain.ActionIncome && transactionType != domain.ActionExpense {
		return r.showTransactionTypeSelection(ctx, callback.ChatID, callback.MessageID)
	}
	if _, err := r.dialogs.Start(ctx, callback.ChatID, callback.Account.UserID, domain.FlowTransaction, domain.StepTransactionAmount, map[string]string{
		"type": transactionType,
	}); err != nil {
		return err
	}
	return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
		fmt.Sprintf("<b>%s</b>\n\nВведите сумму, например <code>1250.50</code>:", operationLabel(transactionType)),
		ports.NewKeyboard(ports.ButtonRow(ports.CallbackButton("❌ Отмена", "fin:tx:cancel"))))
}

func (r *Router) startEditTransaction(ctx context.Context, callback callbackContext, id int) error {
	if id <= 0 {
		return r.showFinanceMenu(ctx, callback.ChatID, callback.MessageID)
	}
	transaction, err := r.platform.GetTransaction(ctx, callback.Account.UserID, id)
	if err != nil {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, mapPlatformError(err), sectionBack("fin:tx:all", "Операции"))
	}
	data := map[string]string{
		"type":        transaction.Type,
		"category":    transaction.Category,
		"description": transaction.Description,
		"tx_id":       strconv.Itoa(transaction.ID),
	}
	if _, err := r.dialogs.Start(ctx, callback.ChatID, callback.Account.UserID, domain.FlowTransaction, domain.StepTransactionAmount, data); err != nil {
		return err
	}
	return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
		fmt.Sprintf("<b>Редактирование #%d</b>\n\nТекущая сумма: %s ₽.\nВведите новую сумму:", transaction.ID, formatAmount(transaction.Amount)),
		ports.NewKeyboard(ports.ButtonRow(ports.CallbackButton("❌ Отмена", "fin:tx:cancel"))))
}

func (r *Router) handleTransactionText(ctx context.Context, message messageContext, state domain.DialogState) error {
	switch state.Step {
	case domain.StepTransactionAmount:
		amount, err := parseAmount(message.Text)
		if err != nil {
			return r.reply(ctx, message.ChatID, "Сумма должна быть положительным числом, например <code>1250.50</code>.", nil)
		}
		state.Data["amount"] = strconv.FormatFloat(amount, 'f', 2, 64)
		state.Step = domain.StepTransactionCategory
		if _, err := r.dialogs.Update(ctx, state, domain.StepTransactionCategory, state.Data); err != nil {
			return err
		}
		return r.showTransactionCategories(ctx, message.ChatID, message.Account, state)
	case domain.StepTransactionCategory:
		category := strings.TrimSpace(message.Text)
		if len([]rune(category)) < 1 || len([]rune(category)) > 60 {
			return r.reply(ctx, message.ChatID, "Название категории должно быть от 1 до 60 символов.", nil)
		}
		state.Data["category"] = category
		state.Step = domain.StepTransactionNote
		if _, err := r.dialogs.Update(ctx, state, domain.StepTransactionNote, state.Data); err != nil {
			return err
		}
		return r.askTransactionNote(ctx, message.ChatID)
	case domain.StepTransactionNote:
		description := strings.TrimSpace(message.Text)
		if len([]rune(description)) > 500 {
			return r.reply(ctx, message.ChatID, "Описание не должно превышать 500 символов.", nil)
		}
		state.Data["description"] = description
		return r.showTransactionDraft(ctx, message.ChatID, state)
	default:
		return r.reply(ctx, message.ChatID, "Используйте кнопки в сообщении выше.", nil)
	}
}

func (r *Router) showTransactionCategories(ctx context.Context, chatID int64, account domain.TelegramAccount, state domain.DialogState) error {
	keyboard := r.categoryKeyboard(ctx, account.UserID, "fin:tx:cat", true)
	_, err := r.telegram.SendMessage(ctx, chatID,
		fmt.Sprintf("<b>%s · %s ₽</b>\n\nВыберите категорию или отправьте своё название сообщением:", operationLabel(state.Get("type")), formatAmount(parseAmountMust(state.Get("amount")))),
		keyboard)
	return err
}

func (r *Router) selectTransactionCategory(ctx context.Context, callback callbackContext, value string) error {
	state, active := r.dialogs.Get(ctx, callback.ChatID)
	if !active || state.Flow != domain.FlowTransaction || state.UserID != callback.Account.UserID {
		return r.showFinanceMenu(ctx, callback.ChatID, callback.MessageID)
	}
	if state.Step != domain.StepTransactionCategory {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, "Категория уже выбрана.", nil)
	}
	if value == "custom" {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
			"Отправьте название категории сообщением.", nil)
	}
	categories := r.categories(ctx, callback.Account.UserID)
	index, err := strconv.Atoi(value)
	if err != nil || index < 0 || index >= len(categories) {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, "Категория не найдена. Выберите из списка.", nil)
	}
	state.Data["category"] = categories[index]
	state.Step = domain.StepTransactionNote
	if _, err := r.dialogs.Update(ctx, state, domain.StepTransactionNote, state.Data); err != nil {
		return err
	}
	return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
		fmt.Sprintf("<b>Категория:</b> %s\n\nДобавьте описание операции или пропустите этот шаг.", escapeHTML(categories[index])),
		ports.NewKeyboard(
			ports.ButtonRow(
				ports.CallbackButton("⏭ Пропустить", "fin:tx:skipnote"),
				ports.CallbackButton("❌ Отмена", "fin:tx:cancel"),
			),
		))
}

func (r *Router) askTransactionNote(ctx context.Context, chatID int64) error {
	_, err := r.telegram.SendMessage(ctx, chatID,
		"Добавьте описание операции или пропустите этот шаг.",
		ports.NewKeyboard(
			ports.ButtonRow(
				ports.CallbackButton("⏭ Пропустить", "fin:tx:skipnote"),
				ports.CallbackButton("❌ Отмена", "fin:tx:cancel"),
			),
		))
	return err
}

func (r *Router) confirmTransactionDraft(ctx context.Context, callback callbackContext, description string) error {
	state, active := r.dialogs.Get(ctx, callback.ChatID)
	if !active || state.Flow != domain.FlowTransaction || state.UserID != callback.Account.UserID {
		return r.showFinanceMenu(ctx, callback.ChatID, callback.MessageID)
	}
	if description != "" {
		state.Data["description"] = description
	}
	state.Data["description"] = strings.TrimSpace(state.Get("description"))
	if _, err := r.dialogs.Update(ctx, state, domain.StepTransactionConfirm, state.Data); err != nil {
		return err
	}
	return r.editOrSend(ctx, callback.ChatID, callback.MessageID, r.transactionDraftText(state), r.confirmKeyboard("fin:tx:confirm", "fin:tx:cancel"))
}

func (r *Router) showTransactionDraft(ctx context.Context, chatID int64, state domain.DialogState) error {
	state.Step = domain.StepTransactionConfirm
	if _, err := r.dialogs.Update(ctx, state, domain.StepTransactionConfirm, state.Data); err != nil {
		return err
	}
	_, err := r.telegram.SendMessage(ctx, chatID, r.transactionDraftText(state), r.confirmKeyboard("fin:tx:confirm", "fin:tx:cancel"))
	return err
}

func (r *Router) transactionDraftText(state domain.DialogState) string {
	var builder strings.Builder
	builder.WriteString("<b>Проверьте операцию</b>\n\n")
	builder.WriteString(fmt.Sprintf("Тип: <b>%s</b>\n", operationLabel(state.Get("type"))))
	builder.WriteString(fmt.Sprintf("Сумма: <b>%s ₽</b>\n", formatAmount(parseAmountMust(state.Get("amount")))))
	builder.WriteString(fmt.Sprintf("Категория: %s\n", escapeHTML(state.Get("category"))))
	if description := state.Get("description"); description != "" {
		builder.WriteString(fmt.Sprintf("Описание: %s\n", escapeHTML(description)))
	}
	if state.Get("tx_id") != "" {
		builder.WriteString("\nПосле подтверждения операция будет обновлена.")
	} else {
		builder.WriteString("\nПосле подтверждения операция будет сохранена в finance-service.")
	}
	return builder.String()
}

func (r *Router) confirmTransaction(ctx context.Context, callback callbackContext) error {
	state, active := r.dialogs.Get(ctx, callback.ChatID)
	if !active || state.Flow != domain.FlowTransaction || state.UserID != callback.Account.UserID || state.Step != domain.StepTransactionConfirm {
		return r.showFinanceMenu(ctx, callback.ChatID, callback.MessageID)
	}
	amount := parseAmountMust(state.Get("amount"))
	category := state.Get("category")
	description := state.Get("description")
	transactionType := state.Get("type")
	if amount <= 0 || category == "" || (transactionType != domain.ActionIncome && transactionType != domain.ActionExpense) {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, "Черновик операции повреждён. Начните заново.", sectionBack("fin:tx:all", "Операции"))
	}

	if err := r.dialogs.Clear(ctx, callback.ChatID); err != nil {
		r.logger.Warn("clear dialog state")
	}

	txID := parseID(state.Get("tx_id"))
	var (
		transaction domain.Transaction
		err         error
	)
	if txID > 0 {
		transaction, err = r.platform.UpdateTransaction(ctx, callback.Account.UserID, txID, transactionType, amount, category, description)
	} else {
		transaction, err = r.platform.CreateTransaction(ctx, callback.Account.UserID, transactionType, amount, category, description)
	}
	if err != nil {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
			"Не удалось сохранить операцию. "+mapPlatformError(err), sectionBack("fin:tx:all", "Операции"))
	}

	action := "создана"
	if txID > 0 {
		action = "обновлена"
	}
	return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
		fmt.Sprintf("<b>Операция %s.</b>\n\n#%d · %s · %s ₽ · %s", action, transaction.ID, operationLabel(transaction.Type), formatAmount(transaction.Amount), escapeHTML(transaction.Category)),
		ports.NewKeyboard(
			ports.ButtonRow(
				ports.CallbackButton("👁 Открыть", fmt.Sprintf("fin:tx:view:%d", transaction.ID)),
				ports.CallbackButton("📋 Все операции", "fin:tx:all"),
			),
			ports.ButtonRow(
				ports.CallbackButton("➕ Ещё", "fin:tx:add"),
				ports.CallbackButton("🏠 Меню", "menu"),
			),
		))
}

func (r *Router) showTransaction(ctx context.Context, callback callbackContext, id int) error {
	if id <= 0 {
		return r.showFinanceMenu(ctx, callback.ChatID, callback.MessageID)
	}
	transaction, err := r.platform.GetTransaction(ctx, callback.Account.UserID, id)
	if err != nil {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, mapPlatformError(err), sectionBack("fin:tx:all", "Операции"))
	}
	text := fmt.Sprintf(
		"<b>Операция #%d</b>\n\nТип: <b>%s</b>\nСумма: <b>%s ₽</b>\nКатегория: %s\nДата: %s",
		transaction.ID,
		operationLabel(transaction.Type),
		formatAmount(transaction.Amount),
		escapeHTML(transaction.Category),
		formatDate(transaction.CreatedAt),
	)
	if transaction.Description != "" {
		text += fmt.Sprintf("\nОписание: %s", escapeHTML(transaction.Description))
	}
	return r.editOrSend(ctx, callback.ChatID, callback.MessageID, text,
		ports.NewKeyboard(
			ports.ButtonRow(
				ports.CallbackButton("✏️ Изменить", fmt.Sprintf("fin:tx:edit:%d", transaction.ID)),
				ports.CallbackButton("🗑 Удалить", fmt.Sprintf("fin:tx:del:%d", transaction.ID)),
			),
			ports.ButtonRow(
				ports.CallbackButton("⬅️ Операции", "fin:tx:all"),
				ports.CallbackButton("🏠 Меню", "menu"),
			),
		))
}

func (r *Router) handleDeleteTransaction(ctx context.Context, callback callbackContext, parts []string) error {
	id := parseID(argAt(parts, 3))
	if id <= 0 {
		return r.showFinanceMenu(ctx, callback.ChatID, callback.MessageID)
	}
	if argAt(parts, 4) != "yes" {
		transaction, err := r.platform.GetTransaction(ctx, callback.Account.UserID, id)
		if err != nil {
			return r.editOrSend(ctx, callback.ChatID, callback.MessageID, mapPlatformError(err), sectionBack("fin:tx:all", "Операции"))
		}
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
			fmt.Sprintf("<b>Удалить операцию #%d?</b>\n\n%s · %s ₽ · %s\n\nДействие необратимо.", transaction.ID, operationLabel(transaction.Type), formatAmount(transaction.Amount), escapeHTML(transaction.Category)),
			ports.NewKeyboard(
				ports.ButtonRow(
					ports.CallbackButton("🗑 Да, удалить", fmt.Sprintf("fin:tx:del:%d:yes", id)),
					ports.CallbackButton("⬅️ Отмена", fmt.Sprintf("fin:tx:view:%d", id)),
				),
			))
	}
	if err := r.platform.DeleteTransaction(ctx, callback.Account.UserID, id); err != nil {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, mapPlatformError(err), sectionBack("fin:tx:all", "Операции"))
	}
	return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
		fmt.Sprintf("<b>Операция #%d удалена.</b>", id),
		sectionBack("fin:tx:all", "Операции"))
}

func (r *Router) showFilterMenu(ctx context.Context, chatID int64, messageID int) error {
	return r.editOrSend(ctx, chatID, messageID,
		"<b>Фильтры операций</b>\n\nТип, категория и период применяются к списку операций.",
		ports.NewKeyboard(
			ports.ButtonRow(
				ports.CallbackButton("Все", "fin:tx:ft:all"),
				ports.CallbackButton("Доходы", "fin:tx:ft:income"),
				ports.CallbackButton("Расходы", "fin:tx:ft:expense"),
			),
			ports.ButtonRow(
				ports.CallbackButton("📂 По категории", "fin:tx:fcat"),
				ports.CallbackButton("📅 По датам", "fin:tx:fdate"),
			),
			ports.ButtonRow(
				ports.CallbackButton("⬅️ К операциям", "fin:tx:all"),
				ports.CallbackButton("🏠 Меню", "menu"),
			),
		))
}

func (r *Router) startDateFilter(ctx context.Context, callback callbackContext) error {
	filter := r.currentFilter(ctx, callback.ChatID, callback.Account.UserID)
	data := map[string]string{
		"type":     filter.Type,
		"category": filter.Category,
		"from":     filter.From,
		"to":       filter.To,
	}
	if _, err := r.dialogs.Start(ctx, callback.ChatID, callback.Account.UserID, domain.FlowFilter, domain.StepFilterDates, data); err != nil {
		return err
	}
	return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
		"Введите период двумя датами: <code>2026-01-01 2026-01-31</code>.",
		ports.NewKeyboard(ports.ButtonRow(ports.CallbackButton("⬅️ Отмена", "fin:tx:all"))))
}

func (r *Router) showFilterCategories(ctx context.Context, callback callbackContext) error {
	keyboard := r.categoryKeyboard(ctx, callback.Account.UserID, "fin:tx:fc", false)
	return r.editOrSend(ctx, callback.ChatID, callback.MessageID, "Выберите категорию для фильтра:", keyboard)
}

func (r *Router) applyFilterCategory(ctx context.Context, callback callbackContext, value string) error {
	categories := r.categories(ctx, callback.Account.UserID)
	index, err := strconv.Atoi(value)
	if err != nil || index < 0 || index >= len(categories) {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, "Категория не найдена.", sectionBack("fin:tx:filter", "Фильтры"))
	}
	filter := r.currentFilter(ctx, callback.ChatID, callback.Account.UserID)
	filter.Category = categories[index]
	if err := r.saveFilter(ctx, callback, filter); err != nil {
		return err
	}
	return r.showTransactions(ctx, callback.ChatID, callback.MessageID, callback.Account, 0)
}

func (r *Router) applyFilterType(ctx context.Context, callback callbackContext, transactionType string) error {
	filter := r.currentFilter(ctx, callback.ChatID, callback.Account.UserID)
	switch transactionType {
	case domain.ActionIncome, domain.ActionExpense:
		filter.Type = transactionType
	default:
		filter.Type = ""
	}
	if err := r.saveFilter(ctx, callback, filter); err != nil {
		return err
	}
	return r.showTransactions(ctx, callback.ChatID, callback.MessageID, callback.Account, 0)
}

func (r *Router) currentFilter(ctx context.Context, chatID int64, userID int) domain.TransactionFilter {
	filter := domain.TransactionFilter{Limit: 10}
	state, active := r.dialogs.Get(ctx, chatID)
	if !active || state.Flow != domain.FlowFilter || state.UserID != userID {
		return filter
	}
	filter.Type = state.Get("type")
	filter.Category = state.Get("category")
	filter.From = state.Get("from")
	filter.To = state.Get("to")
	return filter
}

func (r *Router) saveFilter(ctx context.Context, callback callbackContext, filter domain.TransactionFilter) error {
	data := map[string]string{
		"type":     filter.Type,
		"category": filter.Category,
		"from":     filter.From,
		"to":       filter.To,
	}
	_, err := r.dialogs.Start(ctx, callback.ChatID, callback.Account.UserID, domain.FlowFilter, "list", data)
	return err
}

func (r *Router) categoryKeyboard(ctx context.Context, userID int, prefix string, custom bool) *ports.Keyboard {
	categories := r.categories(ctx, userID)
	rows := make([][]ports.Button, 0, len(categories)/2+2)
	for index := 0; index < len(categories); index += 2 {
		row := []ports.Button{
			ports.CallbackButton(truncate(categories[index], 28), fmt.Sprintf("%s:%d", prefix, index)),
		}
		if index+1 < len(categories) {
			row = append(row, ports.CallbackButton(truncate(categories[index+1], 28), fmt.Sprintf("%s:%d", prefix, index+1)))
		}
		rows = append(rows, row)
	}
	if custom {
		rows = append(rows, ports.ButtonRow(ports.CallbackButton("✍️ Своя категория", prefix+":custom")))
	}
	rows = append(rows, ports.ButtonRow(
		ports.CallbackButton("⬅️ Назад", "fin:tx:all"),
		ports.CallbackButton("🏠 Меню", "menu"),
	))
	return ports.NewKeyboard(rows...)
}

func (r *Router) categories(ctx context.Context, userID int) []string {
	categories, err := r.platform.GetCategories(ctx, userID)
	if err != nil {
		categories = nil
	}
	seen := make(map[string]bool, len(categories))
	result := make([]string, 0, len(categories)+len(defaultCategories))
	for _, category := range categories {
		category = strings.TrimSpace(category)
		if category == "" || seen[category] {
			continue
		}
		seen[category] = true
		result = append(result, category)
	}
	for _, category := range defaultCategories {
		if !seen[category] {
			seen[category] = true
			result = append(result, category)
		}
	}
	if len(result) > 14 {
		result = result[:14]
	}
	return result
}

func filterDescription(filter domain.TransactionFilter) string {
	parts := make([]string, 0, 3)
	if filter.Type != "" {
		parts = append(parts, "тип: "+strings.ToLower(operationLabel(filter.Type)))
	}
	if filter.Category != "" {
		parts = append(parts, "категория: "+escapeHTML(filter.Category))
	}
	if filter.From != "" || filter.To != "" {
		parts = append(parts, fmt.Sprintf("период: %s — %s", defaultValue(filter.From, "…"), defaultValue(filter.To, "…")))
	}
	if len(parts) == 0 {
		return "Фильтры не заданы.\n"
	}
	return "Фильтры — " + strings.Join(parts, "; ") + "\n"
}

func parseAmount(value string) (float64, error) {
	normalized := strings.TrimSpace(value)
	normalized = strings.ReplaceAll(normalized, " ", "")
	normalized = strings.ReplaceAll(normalized, "\u00a0", "")
	normalized = strings.ReplaceAll(normalized, ",", ".")
	normalized = strings.TrimSuffix(normalized, "₽")
	amount, err := strconv.ParseFloat(normalized, 64)
	if err != nil || amount <= 0 || amount > 1_000_000_000 {
		return 0, fmt.Errorf("invalid amount")
	}
	return math.Round(amount*100) / 100, nil
}

func parseAmountMust(value string) float64 {
	amount, _ := parseAmount(value)
	return amount
}

func parseID(value string) int {
	id, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || id < 0 {
		return 0
	}
	return id
}

func normalizeOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}

func argAt(parts []string, index int) string {
	if index < 0 || index >= len(parts) {
		return ""
	}
	return parts[index]
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func defaultValue(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func (r *Router) confirmKeyboard(confirm, cancel string) *ports.Keyboard {
	return ports.NewKeyboard(
		ports.ButtonRow(
			ports.CallbackButton("✅ Подтвердить", confirm),
			ports.CallbackButton("❌ Отмена", cancel),
		),
	)
}
