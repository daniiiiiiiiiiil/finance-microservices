package telegram

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/ports"
)

var currencyCodes = []string{"RUB", "USD", "EUR", "GBP", "JPY", "CNY", "AED", "TRY"}

func (r *Router) handleCurrencyCallback(ctx context.Context, callback callbackContext, parts []string) error {
	if len(parts) == 1 {
		return r.showCurrencyMenu(ctx, callback.ChatID, callback.MessageID)
	}
	switch parts[1] {
	case "rates":
		base := argAt(parts, 2)
		if base == "" {
			base = "USD"
		}
		return r.showRates(ctx, callback.ChatID, callback.MessageID, callback.Account.UserID, base)
	case "conv":
		return r.startConversion(ctx, callback)
	case "from":
		return r.selectConversionFrom(ctx, callback, argAt(parts, 2))
	case "to":
		return r.selectConversionTo(ctx, callback, argAt(parts, 2))
	case "cancel":
		if err := r.dialogs.Clear(ctx, callback.ChatID); err != nil {
			r.logger.Warn("clear dialog state")
		}
		return r.showCurrencyMenu(ctx, callback.ChatID, callback.MessageID)
	default:
		return r.showCurrencyMenu(ctx, callback.ChatID, callback.MessageID)
	}
}

func (r *Router) showCurrencyMenu(ctx context.Context, chatID int64, messageID int) error {
	r.clearDialog(ctx, chatID)
	rows := [][]ports.Button{
		ports.ButtonRow(
			ports.CallbackButton("🇺🇸 USD", "cur:rates:USD"),
			ports.CallbackButton("🇪🇺 EUR", "cur:rates:EUR"),
		),
		ports.ButtonRow(
			ports.CallbackButton("🇷🇺 RUB", "cur:rates:RUB"),
			ports.CallbackButton("🇬🇧 GBP", "cur:rates:GBP"),
		),
		ports.ButtonRow(ports.CallbackButton("💱 Конвертер", "cur:conv")),
		ports.ButtonRow(ports.CallbackButton("🏠 Главное меню", "menu")),
	}
	return r.editOrSend(ctx, chatID, messageID,
		"<b>Валюты</b>\n\nАктуальные курсы и конвертация через currency-service. Выберите базовую валюту или откройте конвертер:",
		ports.NewKeyboard(rows...))
}

func (r *Router) showRates(ctx context.Context, chatID int64, messageID, userID int, base string) error {
	base = strings.ToUpper(strings.TrimSpace(base))
	rates, err := r.platform.GetRates(ctx, userID, base)
	if err != nil {
		return r.editOrSend(ctx, chatID, messageID, "Не удалось получить курсы. "+mapPlatformError(err), sectionBack("cur", "Валюты"))
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("<b>Курсы к %s</b>\n\n", escapeHTML(rates.Base)))
	keys := make([]string, 0, len(rates.Rates))
	for code := range rates.Rates {
		keys = append(keys, code)
	}
	sort.Strings(keys)
	count := 0
	for _, code := range keys {
		if code == rates.Base {
			continue
		}
		builder.WriteString(fmt.Sprintf("1 %s = <b>%s %s</b>\n", escapeHTML(rates.Base), formatAmount(rates.Rates[code]), escapeHTML(code)))
		count++
		if count >= 10 {
			break
		}
	}
	if !rates.Timestamp.IsZero() {
		builder.WriteString(fmt.Sprintf("\nОбновлено: %s", formatDate(rates.Timestamp)))
	}

	return r.editOrSend(ctx, chatID, messageID, builder.String(),
		ports.NewKeyboard(
			ports.ButtonRow(
				ports.CallbackButton("🔄 Обновить", "cur:rates:"+base),
				ports.CallbackButton("💱 Конвертер", "cur:conv"),
			),
			ports.ButtonRow(
				ports.CallbackButton("💱 Валюты", "cur"),
				ports.CallbackButton("🏠 Меню", "menu"),
			),
		))
}

func (r *Router) startConversion(ctx context.Context, callback callbackContext) error {
	if _, err := r.dialogs.Start(ctx, callback.ChatID, callback.Account.UserID, domain.FlowCurrency, domain.StepCurrencyFrom, map[string]string{}); err != nil {
		return err
	}
	return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
		"<b>Конвертер</b>\n\nВыберите исходную валюту:", r.currencyKeyboard("cur:from", ""))
}

func (r *Router) selectConversionFrom(ctx context.Context, callback callbackContext, code string) error {
	code = strings.ToUpper(code)
	if !isCurrencyCode(code) {
		return r.showCurrencyMenu(ctx, callback.ChatID, callback.MessageID)
	}
	state, active := r.dialogs.Get(ctx, callback.ChatID)
	if !active || state.Flow != domain.FlowCurrency || state.UserID != callback.Account.UserID {
		return r.showCurrencyMenu(ctx, callback.ChatID, callback.MessageID)
	}
	state.Data["from"] = code
	state.Step = domain.StepCurrencyTo
	if _, err := r.dialogs.Update(ctx, state, domain.StepCurrencyTo, state.Data); err != nil {
		return err
	}
	return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
		fmt.Sprintf("<b>Из %s</b>\n\nВыберите целевую валюту:", code), r.currencyKeyboard("cur:to", code))
}

func (r *Router) selectConversionTo(ctx context.Context, callback callbackContext, code string) error {
	code = strings.ToUpper(code)
	if !isCurrencyCode(code) {
		return r.showCurrencyMenu(ctx, callback.ChatID, callback.MessageID)
	}
	state, active := r.dialogs.Get(ctx, callback.ChatID)
	if !active || state.Flow != domain.FlowCurrency || state.UserID != callback.Account.UserID {
		return r.showCurrencyMenu(ctx, callback.ChatID, callback.MessageID)
	}
	if state.Get("from") == code {
		return r.editOrSend(ctx, callback.ChatID, callback.MessageID, "Исходная и целевая валюта должны различаться.", r.currencyKeyboard("cur:to", code))
	}
	state.Data["to"] = code
	state.Step = domain.StepCurrencyAmount
	if _, err := r.dialogs.Update(ctx, state, domain.StepCurrencyAmount, state.Data); err != nil {
		return err
	}
	return r.editOrSend(ctx, callback.ChatID, callback.MessageID,
		fmt.Sprintf("<b>%s → %s</b>\n\nВведите сумму в %s:", state.Get("from"), code, state.Get("from")),
		ports.NewKeyboard(ports.ButtonRow(ports.CallbackButton("Отмена", "cur:cancel"))))
}

func (r *Router) handleCurrencyText(ctx context.Context, message messageContext, state domain.DialogState) error {
	if state.Step != domain.StepCurrencyAmount {
		return r.reply(ctx, message.ChatID, "Используйте кнопки в сообщении выше.", nil)
	}
	amount, err := parseAmount(message.Text)
	if err != nil {
		return r.reply(ctx, message.ChatID, "Введите положительную сумму, например <code>100</code>.", nil)
	}
	from := state.Get("from")
	to := state.Get("to")
	if from == "" || to == "" {
		return r.reply(ctx, message.ChatID, "Не удалось определить валюты. Начните конвертацию заново.", sectionBack("cur", "Валюты"))
	}
	if err := r.dialogs.Clear(ctx, message.ChatID); err != nil {
		r.logger.Warn("clear dialog state")
	}

	conversion, err := r.platform.Convert(ctx, message.Account.UserID, from, to, amount)
	if err != nil {
		return r.reply(ctx, message.ChatID, "Не удалось выполнить конвертацию. "+mapPlatformError(err), sectionBack("cur", "Валюты"))
	}
	text := fmt.Sprintf(
		"<b>%s %s = %s %s</b>\n\nКурс: 1 %s = %s %s",
		formatAmount(conversion.Amount), escapeHTML(conversion.From),
		formatAmount(conversion.Result), escapeHTML(conversion.To),
		escapeHTML(conversion.From), formatAmount(conversion.Rate), escapeHTML(conversion.To),
	)
	return r.reply(ctx, message.ChatID, text,
		ports.NewKeyboard(
			ports.ButtonRow(
				ports.CallbackButton("💱 Ещё", "cur:conv"),
				ports.CallbackButton("💱 Валюты", "cur"),
			),
			ports.ButtonRow(ports.CallbackButton("🏠 Меню", "menu")),
		))
}

func (r *Router) currencyKeyboard(prefix, exclude string) *ports.Keyboard {
	rows := make([][]ports.Button, 0, len(currencyCodes)/2+1)
	row := make([]ports.Button, 0, 2)
	for _, code := range currencyCodes {
		if code == exclude {
			continue
		}
		row = append(row, ports.CallbackButton(code, prefix+":"+code))
		if len(row) == 2 {
			rows = append(rows, row)
			row = make([]ports.Button, 0, 2)
		}
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}
	rows = append(rows, ports.ButtonRow(
		ports.CallbackButton("⬅️ Назад", "cur"),
		ports.CallbackButton("🏠 Меню", "menu"),
	))
	return ports.NewKeyboard(rows...)
}

func isCurrencyCode(code string) bool {
	for _, item := range currencyCodes {
		if item == code {
			return true
		}
	}
	return false
}
