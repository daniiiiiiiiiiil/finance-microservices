package telegram

import (
	"context"
	"errors"
	"strings"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/features/service"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/logger"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

type Router struct {
	telegram  ports.TelegramAPI
	accounts  ports.AccountServiceInterface
	dialogs   ports.DialogServiceInterface
	platform  ports.PlatformClient
	logger    *logger.Logger
	webAppURL string
}

func NewRouter(
	telegram ports.TelegramAPI,
	accounts ports.AccountServiceInterface,
	dialogs ports.DialogServiceInterface,
	platform ports.PlatformClient,
	logger *logger.Logger,
	webAppURL string,
) *Router {
	return &Router{
		telegram:  telegram,
		accounts:  accounts,
		dialogs:   dialogs,
		platform:  platform,
		logger:    logger,
		webAppURL: webAppURL,
	}
}

type messageContext struct {
	ChatID         int64
	TelegramUserID int64
	Username       *string
	Text           string
	MessageID      int
	Location       *models.Location
	Account        domain.TelegramAccount
}

type callbackContext struct {
	ChatID         int64
	TelegramUserID int64
	Username       *string
	CallbackID     string
	Data           string
	MessageID      int
	Account        domain.TelegramAccount
}

func (r *Router) Process(ctx context.Context, update models.Update) error {
	switch {
	case update.CallbackQuery != nil:
		return r.processCallback(ctx, update.CallbackQuery)
	case update.Message != nil:
		return r.processMessage(ctx, update.Message)
	default:
		return nil
	}
}

func (r *Router) processMessage(ctx context.Context, message *models.Message) error {
	if message.From == nil || message.From.IsBot || message.Chat.ID == 0 {
		return nil
	}
	if message.Chat.Type != "" && message.Chat.Type != models.ChatTypePrivate {
		return r.reply(ctx, message.Chat.ID, "Бот работает только в личных сообщениях.", nil)
	}
	text := strings.TrimSpace(message.Text)
	messageCtx := messageContext{
		ChatID:         message.Chat.ID,
		TelegramUserID: message.From.ID,
		Username:       optionalString(message.From.Username),
		Text:           text,
		MessageID:      message.ID,
		Location:       message.Location,
	}

	account, err := r.accounts.Resolve(ctx, messageCtx.TelegramUserID)
	if err != nil {
		if errors.Is(err, service.ErrNotLinked) {
			return r.handleUnlinkedMessage(ctx, messageCtx)
		}
		r.logger.Error("resolve telegram account", zap.Error(err))
		return r.reply(ctx, messageCtx.ChatID, "Не удалось получить доступ к аккаунту. Попробуйте позже.", backToMenu())
	}
	messageCtx.Account = account

	command, arguments := parseCommand(text)
	if command != "" {
		return r.handleCommand(ctx, messageCtx, command, arguments)
	}

	if messageCtx.Location != nil {
		return r.handleLocation(ctx, messageCtx)
	}

	state, active := r.dialogs.Get(ctx, messageCtx.ChatID)
	if len(message.Photo) > 0 {
		if active && state.UserID == account.UserID && state.Flow == domain.FlowShopping && state.Step == domain.StepShoppingImage {
			return r.uploadShoppingImage(ctx, messageCtx, state, message.Photo[len(message.Photo)-1].FileID)
		}
		return r.reply(ctx, messageCtx.ChatID, "Фото можно загрузить из карточки списка покупок.", backToMenu())
	}
	if active && state.UserID == account.UserID {
		return r.handleFlowText(ctx, messageCtx, state)
	}

	if looksLikeLinkCode(text) {
		return r.reply(ctx, messageCtx.ChatID, "Похоже, это код привязки. Отправьте его командой <code>/start КОД</code>.", backToMenu())
	}

	return r.showMenu(ctx, messageCtx.ChatID, 0)
}

func (r *Router) processCallback(ctx context.Context, callback *models.CallbackQuery) error {
	callbackCtx := callbackContext{
		TelegramUserID: callback.From.ID,
		Username:       optionalString(callback.From.Username),
		CallbackID:     callback.ID,
		Data:           callback.Data,
	}
	if callback.Message.Message != nil {
		callbackCtx.ChatID = callback.Message.Message.Chat.ID
		callbackCtx.MessageID = callback.Message.Message.ID
		if callback.Message.Message.Chat.Type != "" && callback.Message.Message.Chat.Type != models.ChatTypePrivate {
			return nil
		}
	} else if callback.Message.InaccessibleMessage != nil {
		callbackCtx.ChatID = callback.Message.InaccessibleMessage.Chat.ID
		callbackCtx.MessageID = callback.Message.InaccessibleMessage.MessageID
	}
	if callbackCtx.ChatID == 0 {
		callbackCtx.ChatID = callback.From.ID
	}

	if err := r.telegram.AnswerCallback(ctx, callback.ID, "", false); err != nil {
		r.logger.Debug("answer callback failed", zap.Error(err))
	}
	if callback.Data == "" {
		return nil
	}

	account, err := r.accounts.Resolve(ctx, callbackCtx.TelegramUserID)
	if err != nil {
		if errors.Is(err, service.ErrNotLinked) {
			return r.editOrSend(ctx, callbackCtx.ChatID, callbackCtx.MessageID,
				"Аккаунт не привязан.\n\nОткройте веб-приложение, раздел «Telegram-бот», нажмите «Подключить Telegram» и отправьте боту <code>/start КОД</code>.",
				backToMenu())
		}
		r.logger.Error("resolve telegram account", zap.Error(err))
		return r.editOrSend(ctx, callbackCtx.ChatID, callbackCtx.MessageID, "Не удалось получить доступ к аккаунту. Попробуйте позже.", backToMenu())
	}
	callbackCtx.Account = account

	return r.routeCallback(ctx, callbackCtx)
}

func (r *Router) routeCallback(ctx context.Context, callback callbackContext) error {
	parts := strings.Split(callback.Data, ":")
	switch parts[0] {
	case "menu":
		return r.showMenu(ctx, callback.ChatID, callback.MessageID)
	case "noop":
		return nil
	case "fin":
		return r.handleFinanceCallback(ctx, callback, parts)
	case "an":
		return r.handleAnalyticsCallback(ctx, callback, parts)
	case "sh":
		return r.handleShoppingCallback(ctx, callback, parts)
	case "ai":
		return r.handleAICallback(ctx, callback, parts)
	case "cur":
		return r.handleCurrencyCallback(ctx, callback, parts)
	case "ntf":
		return r.handleNotificationsCallback(ctx, callback, parts)
	case "set":
		return r.handleSettingsCallback(ctx, callback, parts)
	default:
		return r.showMenu(ctx, callback.ChatID, callback.MessageID)
	}
}

func (r *Router) handleCommand(ctx context.Context, message messageContext, command string, arguments []string) error {
	switch command {
	case "/start":
		return r.handleStart(ctx, message, arguments)
	case "/help":
		return r.handleHelp(ctx, message)
	case "/menu":
		return r.showMenu(ctx, message.ChatID, 0)
	case "/settings":
		return r.showSettings(ctx, message.ChatID, 0, message.Account)
	case "/unlink":
		return r.confirmUnlink(ctx, message.ChatID, 0)
	case "/balance":
		return r.showFinanceDashboard(ctx, message.ChatID, 0, message.Account.UserID)
	default:
		return r.reply(ctx, message.ChatID, "Неизвестная команда. Откройте меню кнопкой ниже.", backToMenu())
	}
}

func (r *Router) handleFlowText(ctx context.Context, message messageContext, state domain.DialogState) error {
	switch state.Flow {
	case domain.FlowTransaction:
		return r.handleTransactionText(ctx, message, state)
	case domain.FlowBudget:
		return r.handleBudgetText(ctx, message, state)
	case domain.FlowShopping:
		return r.handleShoppingText(ctx, message, state)
	case domain.FlowAIPlan:
		return r.handleAIPlanText(ctx, message, state)
	case domain.FlowCurrency:
		return r.handleCurrencyText(ctx, message, state)
	case domain.FlowFilter:
		return r.handleFinanceFilterText(ctx, message, state)
	default:
		if err := r.dialogs.Clear(ctx, message.ChatID); err != nil {
			r.logger.Warn("clear dialog state", zap.Error(err))
		}
		return r.showMenu(ctx, message.ChatID, 0)
	}
}

func (r *Router) handleUnlinkedMessage(ctx context.Context, message messageContext) error {
	command, arguments := parseCommand(message.Text)
	if command == "/start" && len(arguments) > 0 {
		return r.handleLinkCode(ctx, message, arguments[0])
	}
	if command == "/help" {
		return r.reply(ctx, message.ChatID, helpText(), nil)
	}
	if command == "/start" || command == "/menu" {
		return r.reply(ctx, message.ChatID, unlinkedText(), nil)
	}
	return r.reply(ctx, message.ChatID, unlinkedText(), nil)
}

func (r *Router) handleLocation(ctx context.Context, message messageContext) error {
	state, active := r.dialogs.Get(ctx, message.ChatID)
	if active && state.UserID == message.Account.UserID && state.Flow == domain.FlowAIPlan && state.Step == domain.StepAIPlanAddress {
		return r.applyAIPlanLocation(ctx, message, state)
	}
	return r.reply(ctx, message.ChatID,
		"Геолокация получена. Координаты можно использовать в AI-планировщике покупок.",
		ports.NewKeyboard(
			ports.ButtonRow(ports.CallbackButton("🤖 AI-планировщик", "ai"), ports.CallbackButton("🏠 Главное меню", "menu")),
		))
}

func (r *Router) clearDialog(ctx context.Context, chatID int64) {
	if err := r.dialogs.Clear(ctx, chatID); err != nil {
		r.logger.Debug("clear dialog state failed", zap.Error(err))
	}
}

func (r *Router) reply(ctx context.Context, chatID int64, text string, keyboard *ports.Keyboard) error {
	_, err := r.telegram.SendMessage(ctx, chatID, truncateMessage(text), keyboard)
	if err != nil {
		r.logger.Error("send telegram message", zap.Error(err))
		return err
	}
	return nil
}

func (r *Router) editOrSend(ctx context.Context, chatID int64, messageID int, text string, keyboard *ports.Keyboard) error {
	text = truncateMessage(text)
	if messageID > 0 {
		if err := r.telegram.EditMessage(ctx, chatID, messageID, text, keyboard); err == nil {
			return nil
		}
	}
	_, err := r.telegram.SendMessage(ctx, chatID, text, keyboard)
	if err != nil {
		r.logger.Error("send telegram message", zap.Error(err))
		return err
	}
	return nil
}

func truncateMessage(text string) string {
	const maxMessageRunes = 4000
	return truncate(text, maxMessageRunes)
}

func parseCommand(text string) (string, []string) {
	if !strings.HasPrefix(text, "/") {
		return "", nil
	}
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return "", nil
	}
	command := fields[0]
	if index := strings.Index(command, "@"); index > 0 {
		command = command[:index]
	}
	command = strings.ToLower(command)
	return command, fields[1:]
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func looksLikeLinkCode(text string) bool {
	if len(text) != 8 {
		return false
	}
	for _, char := range text {
		if (char < 'A' || char > 'Z') && (char < '0' || char > '9') {
			return false
		}
	}
	return true
}
