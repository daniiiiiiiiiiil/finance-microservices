package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/pagination"
)

func (r *Router) handleShoppingCallback(ctx context.Context, callback callbackContext, parts []string) error {
	if len(parts) == 1 {
		return r.showShoppingMenu(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID)
	}
	switch parts[1] {
	case "list":
		return r.showShoppingList(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID, parseID(argAt(parts, 2)))
	case "create":
		return r.startShoppingCreate(ctx, callback)
	case "v":
		return r.showShoppingItem(ctx, callback, parseID(argAt(parts, 2)))
	case "c":
		return r.toggleShopping(ctx, callback, parseID(argAt(parts, 2)))
	case "e":
		return r.startShoppingEdit(ctx, callback, parseID(argAt(parts, 2)))
	case "d":
		return r.handleDeleteShopping(ctx, callback, parts)
	case "i":
		return r.sendShoppingImage(ctx, callback, parseID(argAt(parts, 2)))
	case "up":
		return r.startShoppingImageUpload(ctx, callback, parseID(argAt(parts, 2)))
	case "skipdesc":
		return r.skipShoppingDescription(ctx, callback)
	case "confirm":
		return r.confirmShopping(ctx, callback)
	case "cancel":
		if err := r.dialogs.Clear(ctx, callback.ChatID); err != nil {
			r.logger.Warn("clear dialog state")
		}
		return r.showShoppingMenu(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID)
	default:
		return r.showShoppingMenu(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID)
	}
}

func (r *Router) showShoppingMenu(ctx context.Context, chatID int64, messageID, userID int) error {
	r.clearDialog(ctx, chatID)
	text := "<b>Списки покупок</b>\n\nСоздавайте списки, отслеживайте прогресс и отмечайте завершение. Все изменения синхронизированы с веб-приложением."
	if total, err := r.platform.ShoppingTotal(ctx, userID); err == nil {
		text += fmt.Sprintf("\n\nВсего списков: <b>%d</b>", total)
	}
	return r.editOrSend(ctx, chatID, messageID, text,
		ports.NewKeyboard(
			ports.ButtonRow(
				ports.CallbackButton("📋 Мои списки", "sh:list:0"),
				ports.CallbackButton("➕ Создать", "sh:create"),
			),
			ports.ButtonRow(ports.CallbackButton("🏠 Главное меню", "menu")),
		))
}

func (r *Router) showShoppingList(ctx context.Context, chatID int64, messageID, userID, offset int) error {
	limit := 5
	offset = normalizeOffset(offset)
	page, err := r.platform.ListShopping(ctx, userID, limit, offset)
	if err != nil {
		return r.editOrSend(ctx, chatID, messageID, "Не удалось загрузить списки покупок. Попробуйте позже.", sectionBack("sh", "Списки покупок"))
	}
	if len(page.Items) == 0 {
		return r.editOrSend(ctx, chatID, messageID,
			"<b>Мои списки</b>\n\nСписков пока нет. Создайте первый список.",
			ports.NewKeyboard(
				ports.ButtonRow(ports.CallbackButton("➕ Создать список", "sh:create")),
				ports.ButtonRow(
					ports.CallbackButton("🛒 Раздел", "sh"),
					ports.CallbackButton("🏠 Меню", "menu"),
				),
			))
	}

	meta := pagination.NewPagination(page.Total, limit, offset)
	var builder strings.Builder
	builder.WriteString("<b>Мои списки</b>\n")
	builder.WriteString(fmt.Sprintf("Страница %d из %d · всего %d\n\n", meta.CurrentPage, maxInt(meta.TotalPages, 1), page.Total))

	rows := make([][]ports.Button, 0, len(page.Items)+4)
	for _, item := range page.Items {
		status := "активен"
		if item.Completed {
			status = "завершён"
		}
		builder.WriteString(fmt.Sprintf(
			"<b>%s</b> · %s\n%s\n%s / %s ₽\n\n",
			escapeHTML(truncate(item.Title, 60)),
			status,
			progressBar(item.AmountNow, item.AmountFinish),
			formatAmount(item.AmountNow),
			formatAmount(item.AmountFinish),
		))
		rows = append(rows, ports.ButtonRow(
			ports.CallbackButton("👁 "+truncate(item.Title, 30), fmt.Sprintf("sh:v:%d", item.ID)),
		))
	}

	navRow := make([]ports.Button, 0, 2)
	if meta.HasPrev {
		navRow = append(navRow, ports.CallbackButton("⬅️ Назад", fmt.Sprintf("sh:list:%d", offset-limit)))
	}
	if meta.HasNext {
		navRow = append(navRow, ports.CallbackButton("Вперёд ➡️", fmt.Sprintf("sh:list:%d", offset+limit)))
	}
	if len(navRow) > 0 {
		rows = append(rows, navRow)
	}
	rows = append(rows, ports.ButtonRow(
		ports.CallbackButton("➕ Создать", "sh:create"),
		ports.CallbackButton("🛒 Раздел", "sh"),
	))
	rows = append(rows, ports.ButtonRow(ports.CallbackButton("🏠 Меню", "menu")))

	return r.editOrSend(ctx, chatID, messageID, builder.String(), ports.NewKeyboard(rows...))
}

func (r *Router) showShoppingItem(ctx context.Context, callback callbackContext, id int) error {
	if id <= 0 {
		return r.showShoppingList(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID, 0)
	}
	item, err := r.platform.GetShopping(ctx, callback.Account.UserID, id)
	if err != nil {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, mapPlatformError(err), sectionBack("sh:list:0", "Списки"))
	}

	status := "активен"
	if item.Completed {
		status = "завершён"
	}
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("<b>%s</b>\n\n", escapeHTML(item.Title)))
	if item.Description != nil && *item.Description != "" {
		builder.WriteString(escapeHTML(*item.Description) + "\n\n")
	}
	builder.WriteString(fmt.Sprintf("%s\n", progressBar(item.AmountNow, item.AmountFinish)))
	builder.WriteString(fmt.Sprintf("Накоплено: <b>%s ₽</b>\n", formatAmount(item.AmountNow)))
	builder.WriteString(fmt.Sprintf("Цель: <b>%s ₽</b>\n", formatAmount(item.AmountFinish)))
	builder.WriteString(fmt.Sprintf("Статус: %s\n", status))
	if item.CompletionDate != nil {
		builder.WriteString(fmt.Sprintf("Планируемая дата: %s\n", formatShortDate(*item.CompletionDate)))
	}

	completeLabel := "Отметить выполненным"
	if item.Completed {
		completeLabel = "Вернуть в работу"
	}
	rows := [][]ports.Button{
		ports.ButtonRow(ports.CallbackButton(completeLabel, fmt.Sprintf("sh:c:%d", item.ID))),
		ports.ButtonRow(
			ports.CallbackButton("✏️ Изменить", fmt.Sprintf("sh:e:%d", item.ID)),
			ports.CallbackButton("🗑 Удалить", fmt.Sprintf("sh:d:%d", item.ID)),
		),
	}
	if item.ImageKey != nil && *item.ImageKey != "" {
		rows = append(rows, ports.ButtonRow(
			ports.CallbackButton("🖼 Изображение", fmt.Sprintf("sh:i:%d", item.ID)),
			ports.CallbackButton("Загрузить фото", fmt.Sprintf("sh:up:%d", item.ID)),
		))
	} else {
		rows = append(rows, ports.ButtonRow(
			ports.CallbackButton("Загрузить фото", fmt.Sprintf("sh:up:%d", item.ID)),
		))
	}
	rows = append(rows, ports.ButtonRow(
		ports.CallbackButton("⬅️ Списки", "sh:list:0"),
		ports.CallbackButton("🏠 Меню", "menu"),
	))
	return r.editOrSend(ctx, callback.ChatID, callback.MessageID, builder.String(), ports.NewKeyboard(rows...))
}

func (r *Router) startShoppingCreate(ctx context.Context, callback callbackContext) error {
	if _, err := r.dialogs.Start(ctx, callback.ChatID, callback.Account.UserID, domain.FlowShopping, domain.StepShoppingTitle, map[string]string{}); err != nil {
		return err
	}
	return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
		"<b>Новый список</b>\n\nВведите название списка:",
		ports.NewKeyboard(ports.ButtonRow(ports.CallbackButton("Отмена", "sh:cancel"))))
}

func (r *Router) startShoppingEdit(ctx context.Context, callback callbackContext, id int) error {
	if id <= 0 {
		return r.showShoppingList(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID, 0)
	}
	item, err := r.platform.GetShopping(ctx, callback.Account.UserID, id)
	if err != nil {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, mapPlatformError(err), sectionBack("sh:list:0", "Списки"))
	}
	data := map[string]string{
		"shopping_id":   strconv.Itoa(item.ID),
		"title":         item.Title,
		"amount_now":    strconv.FormatFloat(item.AmountNow, 'f', 2, 64),
		"amount_finish": strconv.FormatFloat(item.AmountFinish, 'f', 2, 64),
		"completed":     strconv.FormatBool(item.Completed),
	}
	if item.Description != nil {
		data["description"] = *item.Description
	}
	if _, err := r.dialogs.Start(ctx, callback.ChatID, callback.Account.UserID, domain.FlowShopping, domain.StepShoppingTitle, data); err != nil {
		return err
	}
	return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
		fmt.Sprintf("<b>Редактирование</b>\n\nТекущее название: %s\nВведите новое название:", escapeHTML(item.Title)),
		ports.NewKeyboard(ports.ButtonRow(ports.CallbackButton("Отмена", "sh:cancel"))))
}

func (r *Router) handleShoppingText(ctx context.Context, message messageContext, state domain.DialogState) error {
	switch state.Step {
	case domain.StepShoppingTitle:
		title := strings.TrimSpace(message.Text)
		if len([]rune(title)) < 1 || len([]rune(title)) >= 200 {
			return r.reply(ctx, message.ChatID, "Название должно быть от 1 до 199 символов.", nil)
		}
		state.Data["title"] = title
		state.Step = domain.StepShoppingDescription
		if _, err := r.dialogs.Update(ctx, state, domain.StepShoppingDescription, state.Data); err != nil {
			return err
		}
		return r.askShoppingDescription(ctx, message.ChatID)
	case domain.StepShoppingDescription:
		description := strings.TrimSpace(message.Text)
		if len([]rune(description)) >= 1000 {
			return r.reply(ctx, message.ChatID, "Описание должно быть короче 1000 символов.", nil)
		}
		state.Data["description"] = description
		state.Step = domain.StepShoppingAmountNow
		if _, err := r.dialogs.Update(ctx, state, domain.StepShoppingAmountNow, state.Data); err != nil {
			return err
		}
		return r.reply(ctx, message.ChatID,
			"Сколько уже накоплено? Введите сумму в рублях (можно 0):",
			ports.NewKeyboard(ports.ButtonRow(ports.CallbackButton("Отмена", "sh:cancel"))))
	case domain.StepShoppingAmountNow:
		amount, err := parseNonNegativeAmount(message.Text)
		if err != nil {
			return r.reply(ctx, message.ChatID, "Введите неотрицательную сумму, например <code>1500</code> или <code>0</code>.", nil)
		}
		state.Data["amount_now"] = strconv.FormatFloat(amount, 'f', 2, 64)
		state.Step = domain.StepShoppingAmountGoal
		if _, err := r.dialogs.Update(ctx, state, domain.StepShoppingAmountGoal, state.Data); err != nil {
			return err
		}
		return r.reply(ctx, message.ChatID,
			"Введите целевую сумму списка в рублях:",
			ports.NewKeyboard(ports.ButtonRow(ports.CallbackButton("Отмена", "sh:cancel"))))
	case domain.StepShoppingAmountGoal:
		amount, err := parseAmount(message.Text)
		if err != nil {
			return r.reply(ctx, message.ChatID, "Целевая сумма должна быть положительным числом.", nil)
		}
		state.Data["amount_finish"] = strconv.FormatFloat(amount, 'f', 2, 64)
		state.Step = domain.StepShoppingConfirm
		if _, err := r.dialogs.Update(ctx, state, domain.StepShoppingConfirm, state.Data); err != nil {
			return err
		}
		return r.reply(ctx, message.ChatID, r.shoppingDraftText(state), r.confirmKeyboard("sh:confirm", "sh:cancel"))
	default:
		return r.reply(ctx, message.ChatID, "Используйте кнопки в сообщении выше.", nil)
	}
}

func (r *Router) askShoppingDescription(ctx context.Context, chatID int64) error {
	_, err := r.telegram.SendMessage(ctx, chatID,
		"Добавьте описание списка или пропустите этот шаг.",
		ports.NewKeyboard(
			ports.ButtonRow(
				ports.CallbackButton("Пропустить", "sh:skipdesc"),
				ports.CallbackButton("Отмена", "sh:cancel"),
			),
		))
	return err
}

func (r *Router) skipShoppingDescription(ctx context.Context, callback callbackContext) error {
	state, active := r.dialogs.Get(ctx, callback.ChatID)
	if !active || state.Flow != domain.FlowShopping || state.UserID != callback.Account.UserID {
		return r.showShoppingMenu(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID)
	}
	if state.Step != domain.StepShoppingDescription {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, "Описание уже обработано.", nil)
	}
	state.Data["description"] = ""
	state.Step = domain.StepShoppingAmountNow
	if _, err := r.dialogs.Update(ctx, state, domain.StepShoppingAmountNow, state.Data); err != nil {
		return err
	}
	return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
		"Сколько уже накоплено? Введите сумму в рублях (можно 0):",
		ports.NewKeyboard(ports.ButtonRow(ports.CallbackButton("Отмена", "sh:cancel"))))
}

func (r *Router) shoppingDraftText(state domain.DialogState) string {
	var builder strings.Builder
	builder.WriteString("<b>Проверьте список</b>\n\n")
	builder.WriteString(fmt.Sprintf("Название: <b>%s</b>\n", escapeHTML(state.Get("title"))))
	if description := state.Get("description"); description != "" {
		builder.WriteString(fmt.Sprintf("Описание: %s\n", escapeHTML(description)))
	}
	builder.WriteString(fmt.Sprintf("Накоплено: %s ₽\n", formatAmount(parseAmountMust(state.Get("amount_now")))))
	builder.WriteString(fmt.Sprintf("Цель: <b>%s ₽</b>\n", formatAmount(parseAmountMust(state.Get("amount_finish")))))
	if state.Get("shopping_id") != "" {
		builder.WriteString("\nПосле подтверждения список будет обновлён.")
	} else {
		builder.WriteString("\nПосле подтверждения список будет создан.")
	}
	return builder.String()
}

func (r *Router) confirmShopping(ctx context.Context, callback callbackContext) error {
	state, active := r.dialogs.Get(ctx, callback.ChatID)
	if !active || state.Flow != domain.FlowShopping || state.UserID != callback.Account.UserID || state.Step != domain.StepShoppingConfirm {
		return r.showShoppingMenu(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID)
	}
	title := state.Get("title")
	amountNow := parseAmountMust(state.Get("amount_now"))
	amountFinish := parseAmountMust(state.Get("amount_finish"))
	if title == "" || amountFinish <= 0 {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, "Черновик списка повреждён. Начните заново.", sectionBack("sh:list:0", "Списки"))
	}

	id := parseID(state.Get("shopping_id"))
	description := state.Get("description")
	var descriptionPtr *string
	if description != "" {
		descriptionPtr = &description
	}
	input := domain.ShoppingInput{
		Title:        title,
		Description:  descriptionPtr,
		AmountNow:    amountNow,
		AmountFinish: amountFinish,
		Completed:    state.Get("completed") == "true",
	}

	if err := r.dialogs.Clear(ctx, callback.ChatID); err != nil {
		r.logger.Warn("clear dialog state")
	}

	var (
		item domain.ShoppingItem
		err  error
	)
	if id > 0 {
		existing, getErr := r.platform.GetShopping(ctx, callback.Account.UserID, id)
		if getErr != nil {
			return r.editOrSend(ctx, callback.ChatID, callback.MessageID, mapPlatformError(getErr), sectionBack("sh:list:0", "Списки"))
		}
		item, err = r.platform.UpdateShopping(ctx, callback.Account.UserID, existing, input)
	} else {
		item, err = r.platform.CreateShopping(ctx, callback.Account.UserID, input)
	}
	if err != nil {
		message := "Не удалось сохранить список. " + mapPlatformError(err)
		if id == 0 && amountNow <= 0 {
			message += "\n\nСервис списков требует положительную сумму накопления при редактировании."
		}
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, message, sectionBack("sh:list:0", "Списки"))
	}

	return r.showShoppingItem(ctx, callbackContext{
		ChatID:     callback.ChatID,
		MessageID:  callback.MessageID,
		CallbackID: callback.CallbackID,
		Account:    callback.Account,
	}, item.ID)
}

func (r *Router) toggleShopping(ctx context.Context, callback callbackContext, id int) error {
	if id <= 0 {
		return r.showShoppingList(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID, 0)
	}
	item, err := r.platform.GetShopping(ctx, callback.Account.UserID, id)
	if err != nil {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, mapPlatformError(err), sectionBack("sh:list:0", "Списки"))
	}
	if err := r.platform.CompleteShopping(ctx, callback.Account.UserID, id, !item.Completed); err != nil {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, mapPlatformError(err), sectionBack("sh:list:0", "Списки"))
	}
	return r.showShoppingItem(ctx, callback, id)
}

func (r *Router) sendShoppingImage(ctx context.Context, callback callbackContext, id int) error {
	if id <= 0 {
		return r.showShoppingList(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID, 0)
	}
	image, err := r.platform.ShoppingImage(ctx, callback.Account.UserID, id)
	if err != nil {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, "Не удалось получить изображение списка.", sectionBack(fmt.Sprintf("sh:v:%d", id), "К списку"))
	}
	if len(image.Data) == 0 {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, "У этого списка нет изображения.", sectionBack(fmt.Sprintf("sh:v:%d", id), "К списку"))
	}
	filename := image.Filename
	if filename == "" {
		filename = "shopping.jpg"
	}
	if err := r.telegram.SendPhoto(ctx, callback.ChatID, image.Data, filename, "Изображение списка"); err != nil {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, "Не удалось отправить изображение.", sectionBack(fmt.Sprintf("sh:v:%d", id), "К списку"))
	}
	return nil
}

func (r *Router) startShoppingImageUpload(ctx context.Context, callback callbackContext, id int) error {
	if id <= 0 {
		return r.showShoppingList(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID, 0)
	}
	item, err := r.platform.GetShopping(ctx, callback.Account.UserID, id)
	if err != nil {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, mapPlatformError(err), sectionBack("sh:list:0", "Списки"))
	}
	data := map[string]string{
		"shopping_id":   strconv.Itoa(item.ID),
		"title":         item.Title,
		"amount_now":    strconv.FormatFloat(item.AmountNow, 'f', 2, 64),
		"amount_finish": strconv.FormatFloat(item.AmountFinish, 'f', 2, 64),
		"completed":     strconv.FormatBool(item.Completed),
	}
	if item.Description != nil {
		data["description"] = *item.Description
	}
	if _, err := r.dialogs.Start(ctx, callback.ChatID, callback.Account.UserID, domain.FlowShopping, domain.StepShoppingImage, data); err != nil {
		return err
	}
	return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
		"<b>Загрузка изображения</b>\n\nОтправьте фотографию в этот чат. Поддерживаются JPG, PNG и WEBP до 10 МБ.",
		ports.NewKeyboard(ports.ButtonRow(ports.CallbackButton("Отмена", "sh:cancel"))))
}

func (r *Router) uploadShoppingImage(ctx context.Context, message messageContext, state domain.DialogState, fileID string) error {
	id := parseID(state.Get("shopping_id"))
	if id <= 0 {
		if err := r.dialogs.Clear(ctx, message.ChatID); err != nil {
			r.logger.Warn("clear dialog state")
		}
		return r.reply(ctx, message.ChatID, "Не удалось определить список. Начните загрузку из карточки списка.", sectionBack("sh:list:0", "Списки"))
	}
	if err := r.telegram.SendChatAction(ctx, message.ChatID, "upload_photo"); err != nil {
		r.logger.Debug("send chat action failed")
	}
	data, err := r.telegram.DownloadFile(ctx, fileID)
	if err != nil {
		return r.reply(ctx, message.ChatID, "Не удалось скачать фотографию. Попробуйте ещё раз.", ports.NewKeyboard(ports.ButtonRow(ports.CallbackButton("❌ Отмена", "sh:cancel"))))
	}
	if len(data) == 0 {
		return r.reply(ctx, message.ChatID, "Фотография пустая. Попробуйте другую.", nil)
	}
	description := state.Get("description")
	var descriptionPtr *string
	if description != "" {
		descriptionPtr = &description
	}
	input := domain.ShoppingInput{
		Title:        state.Get("title"),
		Description:  descriptionPtr,
		AmountNow:    parseAmountMust(state.Get("amount_now")),
		AmountFinish: parseAmountMust(state.Get("amount_finish")),
		Completed:    state.Get("completed") == "true",
		ImageData:    data,
		Filename:     "photo.jpg",
	}
	existing, err := r.platform.GetShopping(ctx, message.Account.UserID, id)
	if err != nil {
		return r.reply(ctx, message.ChatID, mapPlatformError(err), sectionBack("sh:list:0", "Списки"))
	}
	updated, err := r.platform.UpdateShopping(ctx, message.Account.UserID, existing, input)
	if err != nil {
		return r.reply(ctx, message.ChatID, "Не удалось загрузить изображение. "+mapPlatformError(err), sectionBack(fmt.Sprintf("sh:v:%d", id), "К списку"))
	}
	if err := r.dialogs.Clear(ctx, message.ChatID); err != nil {
		r.logger.Warn("clear dialog state")
	}
	return r.showShoppingItem(ctx, callbackContext{
		ChatID:    message.ChatID,
		MessageID: message.MessageID,
		Account:   message.Account,
	}, updated.ID)
}

func (r *Router) handleDeleteShopping(ctx context.Context, callback callbackContext, parts []string) error {
	id := parseID(argAt(parts, 2))
	if id <= 0 {
		return r.showShoppingList(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID, 0)
	}
	if argAt(parts, 3) != "yes" {
		item, err := r.platform.GetShopping(ctx, callback.Account.UserID, id)
		if err != nil {
			return r.editOrSend(ctx, callback.ChatID, callback.MessageID, mapPlatformError(err), sectionBack("sh:list:0", "Списки"))
		}
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
			fmt.Sprintf("<b>Удалить список?</b>\n\n%s\n%s / %s ₽\n\nДействие необратимо.", escapeHTML(item.Title), formatAmount(item.AmountNow), formatAmount(item.AmountFinish)),
			ports.NewKeyboard(
				ports.ButtonRow(
					ports.CallbackButton("Да, удалить", fmt.Sprintf("sh:d:%d:yes", id)),
					ports.CallbackButton("Отмена", fmt.Sprintf("sh:v:%d", id)),
				),
			))
	}
	if err := r.platform.DeleteShopping(ctx, callback.Account.UserID, id); err != nil {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, mapPlatformError(err), sectionBack("sh:list:0", "Списки"))
	}
	return r.editOrSend(ctx, callback.ChatID, callback.MessageID, "<b>Список удалён.</b>", sectionBack("sh:list:0", "Списки"))
}

func parseNonNegativeAmount(value string) (float64, error) {
	normalized := strings.TrimSpace(value)
	normalized = strings.ReplaceAll(normalized, " ", "")
	normalized = strings.ReplaceAll(normalized, "\u00a0", "")
	normalized = strings.ReplaceAll(normalized, ",", ".")
	normalized = strings.TrimSuffix(normalized, "₽")
	amount, err := strconv.ParseFloat(normalized, 64)
	if err != nil || amount < 0 || amount > 1_000_000_000 {
		return 0, fmt.Errorf("invalid amount")
	}
	return mathRound2(amount), nil
}

func mathRound2(value float64) float64 {
	return float64(int64(value*100+0.5)) / 100
}
