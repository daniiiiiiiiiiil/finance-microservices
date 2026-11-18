package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/pagination"
	"github.com/google/uuid"
)

func (r *Router) handleAICallback(ctx context.Context, callback callbackContext, parts []string) error {
	if len(parts) == 1 {
		return r.showAIMenu(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID)
	}
	switch parts[1] {
	case "new":
		return r.startAIPlan(ctx, callback)
	case "run":
		return r.runAIPlan(ctx, callback)
	case "cancel":
		if err := r.dialogs.Clear(ctx, callback.ChatID); err != nil {
			r.logger.Warn("clear dialog state")
		}
		return r.showAIMenu(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID)
	case "hist":
		return r.showAIPlanHistory(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID, parseID(argAt(parts, 2)))
	case "view":
		return r.showAIPlan(ctx, callback, parseID(argAt(parts, 2)))
	case "d":
		return r.handleDeleteAIPlan(ctx, callback, parts)
	case "map":
		return r.showAIMap(ctx, callback, parseID(argAt(parts, 2)))
	default:
		return r.showAIMenu(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID)
	}
}

func (r *Router) showAIMenu(ctx context.Context, chatID int64, messageID, userID int) error {
	r.clearDialog(ctx, chatID)
	text := "<b>AI-планировщик покупок</b>\n\n" +
		"Опишите, что нужно купить, укажите адрес или отправьте геолокацию — ai-service подберёт магазины, цены, доставку и маршрут.\n\n" +
		"AI-логика выполняется только в ai-service."
	rows := [][]ports.Button{
		ports.ButtonRow(
			ports.CallbackButton("✨ Новый план", "ai:new"),
			ports.CallbackButton("🕘 История", "ai:hist:0"),
		),
	}
	if webApp := r.webAppButton("🗺 Web App карта"); webApp.Text != "" {
		rows = append(rows, ports.ButtonRow(webApp))
	}
	rows = append(rows, ports.ButtonRow(ports.CallbackButton("🏠 Главное меню", "menu")))
	return r.editOrSend(ctx, chatID, messageID, text, ports.NewKeyboard(rows...))
}

func (r *Router) startAIPlan(ctx context.Context, callback callbackContext) error {
	if _, err := r.dialogs.Start(ctx, callback.ChatID, callback.Account.UserID, domain.FlowAIPlan, domain.StepAIPlanText, map[string]string{}); err != nil {
		return err
	}
	return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
		"<b>Новый AI-план</b>\n\nЧто нужно купить? Например: <i>молоко, хлеб, яйца и кофе на неделю</i>",
		ports.NewKeyboard(ports.ButtonRow(ports.CallbackButton("Отмена", "ai:cancel"))))
}

func (r *Router) handleAIPlanText(ctx context.Context, message messageContext, state domain.DialogState) error {
	switch state.Step {
	case domain.StepAIPlanText:
		text := strings.TrimSpace(message.Text)
		if len([]rune(text)) < 3 {
			return r.reply(ctx, message.ChatID, "Опишите покупки чуть подробнее.", nil)
		}
		if len([]rune(text)) > 1000 {
			return r.reply(ctx, message.ChatID, "Описание слишком длинное. Максимум 1000 символов.", nil)
		}
		state.Data["text"] = text
		state.Step = domain.StepAIPlanAddress
		if _, err := r.dialogs.Update(ctx, state, domain.StepAIPlanAddress, state.Data); err != nil {
			return err
		}
		return r.reply(ctx, message.ChatID,
			"Укажите адрес отправления текстом или отправьте геолокацию через скрепку Telegram.",
			ports.NewKeyboard(ports.ButtonRow(ports.CallbackButton("Отмена", "ai:cancel"))))
	case domain.StepAIPlanAddress:
		address := strings.TrimSpace(message.Text)
		if len([]rune(address)) < 3 {
			return r.reply(ctx, message.ChatID, "Адрес слишком короткий. Введите адрес текстом или отправьте геолокацию.", nil)
		}
		if len([]rune(address)) > 300 {
			return r.reply(ctx, message.ChatID, "Адрес слишком длинный. Максимум 300 символов.", nil)
		}
		state.Data["address"] = address
		state.Step = domain.StepAIPlanGeo
		if _, err := r.dialogs.Update(ctx, state, domain.StepAIPlanGeo, state.Data); err != nil {
			return err
		}
		return r.reply(ctx, message.ChatID, r.aiPlanDraftText(state), r.aiPlanKeyboard())
	default:
		return r.reply(ctx, message.ChatID, "Используйте кнопки в сообщении выше.", nil)
	}
}

func (r *Router) applyAIPlanLocation(ctx context.Context, message messageContext, state domain.DialogState) error {
	location := message.Location
	if location == nil {
		return nil
	}
	state.Data["latitude"] = strconv.FormatFloat(location.Latitude, 'f', 6, 64)
	state.Data["longitude"] = strconv.FormatFloat(location.Longitude, 'f', 6, 64)
	state.Data["address"] = fmt.Sprintf("%.6f, %.6f", location.Latitude, location.Longitude)
	state.Step = domain.StepAIPlanGeo
	if _, err := r.dialogs.Update(ctx, state, domain.StepAIPlanGeo, state.Data); err != nil {
		return err
	}
	return r.reply(ctx, message.ChatID, r.aiPlanDraftText(state), r.aiPlanKeyboard())
}

func (r *Router) aiPlanDraftText(state domain.DialogState) string {
	var builder strings.Builder
	builder.WriteString("<b>Проверьте запрос</b>\n\n")
	builder.WriteString(fmt.Sprintf("Покупки: %s\n", escapeHTML(state.Get("text"))))
	builder.WriteString(fmt.Sprintf("Адрес: %s\n", escapeHTML(state.Get("address"))))
	if state.Get("latitude") != "" {
		builder.WriteString("Источник адреса: геолокация Telegram\n")
	}
	builder.WriteString("\nЗапустить AI-планирование?")
	return builder.String()
}

func (r *Router) aiPlanKeyboard() *ports.Keyboard {
	return ports.NewKeyboard(
		ports.ButtonRow(
			ports.CallbackButton("Запустить AI", "ai:run"),
			ports.CallbackButton("Отмена", "ai:cancel"),
		),
	)
}

func (r *Router) runAIPlan(ctx context.Context, callback callbackContext) error {
	state, active := r.dialogs.Get(ctx, callback.ChatID)
	if !active || state.Flow != domain.FlowAIPlan || state.UserID != callback.Account.UserID || state.Step != domain.StepAIPlanGeo {
		return r.showAIMenu(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID)
	}
	text := state.Get("text")
	address := state.Get("address")
	if text == "" || address == "" {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, "Черновик запроса повреждён. Начните заново.", sectionBack("ai", "AI-планировщик"))
	}

	latitude, _ := strconv.ParseFloat(state.Get("latitude"), 64)
	longitude, _ := strconv.ParseFloat(state.Get("longitude"), 64)

	if err := r.dialogs.Clear(ctx, callback.ChatID); err != nil {
		r.logger.Warn("clear dialog state")
	}
	if err := r.telegram.EditMessage(ctx, callback.ChatID, callback.MessageID,
		"<b>AI-планирование запущено…</b>\n\nПодбираю магазины, цены и маршрут. Это может занять до минуты.", nil); err != nil {
		r.logger.Debug("edit message failed")
	}

	requestID := uuid.NewString()
	plan, err := r.platform.PlanShopping(ctx, callback.Account.UserID, domain.AIPlanRequest{
		RequestID:    requestID,
		Text:         text,
		StartAddress: address,
		Latitude:     latitude,
		Longitude:    longitude,
	})
	if err != nil {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
			"AI-планирование не удалось. "+mapPlatformError(err),
			ports.NewKeyboard(
				ports.ButtonRow(
					ports.CallbackButton("🔁 Повторить", "ai:new"),
					ports.CallbackButton("⬅️ Раздел", "ai"),
				),
			))
	}

	planID := r.latestPlanID(ctx, callback.Account.UserID, plan.RequestID)
	if len(plan.Stores) > 0 {
		store := plan.Stores[0]
		if store.Latitude != 0 || store.Longitude != 0 {
			if err := r.telegram.SendLocation(ctx, callback.ChatID, store.Latitude, store.Longitude); err != nil {
				r.logger.Debug("send store location failed")
			}
		}
	}
	return r.editOrSend(ctx, callback.ChatID, callback.MessageID, formatAIPlan(plan), r.aiPlanResultKeyboard(planID))
}

func (r *Router) latestPlanID(ctx context.Context, userID int, requestID string) int {
	page, err := r.platform.ListPlans(ctx, userID, 10, 0)
	if err != nil {
		return 0
	}
	for _, plan := range page.Items {
		if plan.RequestID == requestID {
			return plan.ID
		}
	}
	if len(page.Items) > 0 {
		return page.Items[0].ID
	}
	return 0
}

func (r *Router) aiPlanResultKeyboard(planID int) *ports.Keyboard {
	rows := [][]ports.Button{}
	if planID > 0 {
		rows = append(rows, ports.ButtonRow(
			ports.CallbackButton("👁 План", fmt.Sprintf("ai:view:%d", planID)),
			ports.CallbackButton("🗺 Карта", fmt.Sprintf("ai:map:%d", planID)),
		))
	}
	rows = append(rows, ports.ButtonRow(
		ports.CallbackButton("✨ Новый план", "ai:new"),
		ports.CallbackButton("🕘 История", "ai:hist:0"),
	))
	rows = append(rows, ports.ButtonRow(ports.CallbackButton("🏠 Меню", "menu")))
	return ports.NewKeyboard(rows...)
}

func (r *Router) showAIPlanHistory(ctx context.Context, chatID int64, messageID, userID, offset int) error {
	limit := 5
	offset = normalizeOffset(offset)
	page, err := r.platform.ListPlans(ctx, userID, limit, offset)
	if err != nil {
		return r.editOrSend(ctx, chatID, messageID, "Не удалось загрузить историю планов.", sectionBack("ai", "AI-планировщик"))
	}
	if len(page.Items) == 0 {
		return r.editOrSend(ctx, chatID, messageID,
			"<b>История AI-планов</b>\n\nПланов пока нет.",
			ports.NewKeyboard(
				ports.ButtonRow(ports.CallbackButton("✨ Новый план", "ai:new")),
				ports.ButtonRow(ports.CallbackButton("⬅️ Раздел", "ai"), ports.CallbackButton("🏠 Меню", "menu")),
			))
	}

	meta := pagination.NewPagination(page.Total, limit, offset)
	var builder strings.Builder
	builder.WriteString("<b>История AI-планов</b>\n")
	builder.WriteString(fmt.Sprintf("Страница %d из %d · всего %d\n\n", meta.CurrentPage, maxInt(meta.TotalPages, 1), page.Total))
	rows := make([][]ports.Button, 0, len(page.Items)+4)
	for _, plan := range page.Items {
		builder.WriteString(fmt.Sprintf(
			"<b>#%d</b> %s\n%s\n%s ₽ · %s\n\n",
			plan.ID,
			formatDate(plan.CreatedAt),
			escapeHTML(truncate(plan.Text, 90)),
			formatAmount(plan.TotalCost),
			escapeHTML(plan.Recommended),
		))
		rows = append(rows, ports.ButtonRow(
			ports.CallbackButton(fmt.Sprintf("👁 #%d · %s ₽", plan.ID, formatAmount(plan.TotalCost)), fmt.Sprintf("ai:view:%d", plan.ID)),
			ports.CallbackButton("🗑", fmt.Sprintf("ai:d:%d", plan.ID)),
		))
	}
	navRow := make([]ports.Button, 0, 2)
	if meta.HasPrev {
		navRow = append(navRow, ports.CallbackButton("⬅️ Назад", fmt.Sprintf("ai:hist:%d", offset-limit)))
	}
	if meta.HasNext {
		navRow = append(navRow, ports.CallbackButton("Вперёд ➡️", fmt.Sprintf("ai:hist:%d", offset+limit)))
	}
	if len(navRow) > 0 {
		rows = append(rows, navRow)
	}
	rows = append(rows, ports.ButtonRow(
		ports.CallbackButton("✨ Новый план", "ai:new"),
		ports.CallbackButton("⬅️ Раздел", "ai"),
	))
	rows = append(rows, ports.ButtonRow(ports.CallbackButton("🏠 Меню", "menu")))
	return r.editOrSend(ctx, chatID, messageID, builder.String(), ports.NewKeyboard(rows...))
}

func (r *Router) showAIPlan(ctx context.Context, callback callbackContext, id int) error {
	if id <= 0 {
		return r.showAIPlanHistory(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID, 0)
	}
	plan, err := r.platform.GetPlan(ctx, callback.Account.UserID, id)
	if err != nil {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, mapPlatformError(err), sectionBack("ai:hist:0", "История"))
	}
	rows := [][]ports.Button{
		ports.ButtonRow(ports.CallbackButton("🗺 Карта", fmt.Sprintf("ai:map:%d", id))),
		ports.ButtonRow(
			ports.CallbackButton("🕘 История", "ai:hist:0"),
			ports.CallbackButton("🗑 Удалить", fmt.Sprintf("ai:d:%d", id)),
		),
	}
	rows = append(rows, ports.ButtonRow(
		ports.CallbackButton("✨ Новый план", "ai:new"),
		ports.CallbackButton("🏠 Меню", "menu"),
	))
	return r.editOrSend(ctx, callback.ChatID, callback.MessageID, formatAIPlan(plan), ports.NewKeyboard(rows...))
}

func (r *Router) showAIMap(ctx context.Context, callback callbackContext, id int) error {
	if id <= 0 {
		return r.showAIPlanHistory(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID, 0)
	}
	if _, err := r.platform.GetPlan(ctx, callback.Account.UserID, id); err != nil {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, mapPlatformError(err), sectionBack("ai:hist:0", "История"))
	}
	if r.webAppURL == "" {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
			"Web App не настроен. Укажите TELEGRAM_PUBLIC_URL и TELEGRAM_WEBAPP_PATH.",
			sectionBack("ai:hist:0", "История"))
	}
	url := fmt.Sprintf("%s?plan=%d", r.webAppURL, id)
	mapButton := ports.URLButton("🗺 Открыть интерактивную карту", url)
	if strings.HasPrefix(strings.ToLower(r.webAppURL), "https://") {
		mapButton = ports.WebAppButton("🗺 Открыть интерактивную карту", url)
	}
	rows := [][]ports.Button{
		ports.ButtonRow(mapButton),
		ports.ButtonRow(
			ports.CallbackButton("👁 План", fmt.Sprintf("ai:view:%d", id)),
			ports.CallbackButton("🕘 История", "ai:hist:0"),
		),
		ports.ButtonRow(ports.CallbackButton("🏠 Меню", "menu")),
	}
	return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
		"<b>Карта маршрута</b>\n\nОткройте Web App: магазины, цены и маршрут на интерактивной карте.",
		ports.NewKeyboard(rows...))
}

func (r *Router) handleDeleteAIPlan(ctx context.Context, callback callbackContext, parts []string) error {
	id := parseID(argAt(parts, 2))
	if id <= 0 {
		return r.showAIPlanHistory(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID, 0)
	}
	if argAt(parts, 3) != "yes" {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
			fmt.Sprintf("<b>Удалить AI-план #%d?</b>\n\nДействие необратимо.", id),
			ports.NewKeyboard(
				ports.ButtonRow(
					ports.CallbackButton("🗑 Да, удалить", fmt.Sprintf("ai:d:%d:yes", id)),
					ports.CallbackButton("⬅️ Отмена", fmt.Sprintf("ai:view:%d", id)),
				),
			))
	}
	if err := r.platform.DeletePlan(ctx, callback.Account.UserID, id); err != nil {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, mapPlatformError(err), sectionBack("ai:hist:0", "История"))
	}
	return r.editOrSend(ctx, callback.ChatID, callback.MessageID, "<b>AI-план удалён.</b>", sectionBack("ai:hist:0", "История"))
}

func formatAIPlan(plan domain.AIPlan) string {
	var builder strings.Builder
	builder.WriteString("<b>AI-план покупок</b>\n\n")

	if len(plan.Items) > 0 {
		builder.WriteString("<b>Товары</b>\n")
		limit := len(plan.Items)
		if limit > 10 {
			limit = 10
		}
		for _, item := range plan.Items[:limit] {
			quantity := ""
			if item.Quantity > 0 {
				quantity = fmt.Sprintf(" — %d %s", item.Quantity, item.Unit)
			}
			builder.WriteString("• " + escapeHTML(item.Name) + escapeHTML(quantity) + "\n")
		}
		builder.WriteString("\n")
	}

	if len(plan.Stores) > 0 {
		builder.WriteString("<b>Магазины</b>\n")
		limit := len(plan.Stores)
		if limit > 5 {
			limit = 5
		}
		for index, store := range plan.Stores[:limit] {
			builder.WriteString(fmt.Sprintf("%d. <b>%s</b>\n", index+1, escapeHTML(truncate(store.Name, 60))))
			if store.Address != "" {
				builder.WriteString("   " + escapeHTML(truncate(store.Address, 120)) + "\n")
			}
			if store.Latitude != 0 || store.Longitude != 0 {
				builder.WriteString(fmt.Sprintf("   <a href=\"https://www.openstreetmap.org/?mlat=%.6f&mlon=%.6f#map=16/%.6f/%.6f\">На карте</a>\n",
					store.Latitude, store.Longitude, store.Latitude, store.Longitude))
			}
			prices := store.Prices
			if len(prices) > 4 {
				prices = prices[:4]
			}
			for _, price := range prices {
				stock := ""
				if !price.InStock {
					stock = " · нет в наличии"
				}
				builder.WriteString(fmt.Sprintf("   • %s — %s %s%s\n",
					escapeHTML(price.ItemName), formatAmount(price.Price), escapeHTML(price.Currency), stock))
			}
			if store.Delivery != nil && store.Delivery.Available {
				builder.WriteString(fmt.Sprintf("   🚚 Доставка: %s ₽", formatAmount(store.Delivery.Price)))
				if store.Delivery.DurationMin > 0 {
					builder.WriteString(fmt.Sprintf(" · %d мин", store.Delivery.DurationMin))
				}
				builder.WriteString("\n")
			}
		}
		builder.WriteString("\n")
	}

	builder.WriteString(fmt.Sprintf("<b>Товары:</b> %s ₽\n", formatAmount(plan.Cost.ItemsTotal)))
	builder.WriteString(fmt.Sprintf("<b>Дорога:</b> %s ₽\n", formatAmount(plan.Cost.TravelCost)))
	builder.WriteString(fmt.Sprintf("<b>Доставка:</b> %s ₽\n", formatAmount(plan.Cost.DeliveryCost)))
	builder.WriteString(fmt.Sprintf("<b>Итого:</b> %s ₽\n\n", formatAmount(plan.Cost.Total)))

	if plan.Recommendation.Mode != "" {
		builder.WriteString(fmt.Sprintf("<b>Рекомендация:</b> %s\n", escapeHTML(plan.Recommendation.Reason)))
	}
	if plan.Route != nil {
		builder.WriteString(fmt.Sprintf("<b>Маршрут:</b> %.1f км", plan.Route.DistanceMeters/1000))
		if plan.Route.DurationSeconds > 0 {
			builder.WriteString(fmt.Sprintf(" · ~%.0f мин", plan.Route.DurationSeconds/60))
		}
		builder.WriteString("\n")
	}
	return builder.String()
}
