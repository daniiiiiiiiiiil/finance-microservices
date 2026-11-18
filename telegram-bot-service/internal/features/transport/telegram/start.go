package telegram

import (
	"context"
	"errors"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/features/service"
	"go.uber.org/zap"
)

func (r *Router) handleStart(ctx context.Context, message messageContext, arguments []string) error {
	if len(arguments) > 0 {
		return r.reply(ctx, message.ChatID,
			"Аккаунт уже привязан к этому Telegram.\n\nЕсли нужно привязать другой аккаунт, сначала отвяжите текущий командой /unlink.",
			r.mainMenuKeyboard())
	}
	return r.reply(ctx, message.ChatID,
		"<b>С возвращением!</b>\n\nАккаунт привязан. Выберите раздел:",
		r.mainMenuKeyboard())
}

func (r *Router) handleLinkCode(ctx context.Context, message messageContext, code string) error {
	account, err := r.accounts.ConsumeLinkCode(ctx, code, message.TelegramUserID, message.ChatID, message.Username)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCodeInvalid):
			return r.reply(ctx, message.ChatID, "Код недействителен или истёк. Создайте новый код в настройках профиля.", nil)
		case errors.Is(err, service.ErrCodeUsed):
			return r.reply(ctx, message.ChatID, "Этот код уже был использован. Создайте новый код в настройках профиля.", nil)
		case errors.Is(err, service.ErrAlreadyLinked):
			return r.reply(ctx, message.ChatID,
				"Этот Telegram уже привязан к аккаунту, либо у аккаунта уже есть привязанный Telegram.\n\nЕсли нужен другой аккаунт, сначала выполните /unlink.",
				backToMenu())
		default:
			r.logger.Error("consume link code", zap.Error(err))
			return r.reply(ctx, message.ChatID, "Не удалось привязать аккаунт. Попробуйте позже.", nil)
		}
	}

	r.logger.Info("telegram account linked", zap.Int("user_id", account.UserID))
	return r.reply(ctx, message.ChatID,
		"<b>Аккаунт привязан!</b>\n\nТеперь вам доступны финансы, списки покупок, AI-планировщик, аналитика и валюты.",
		r.mainMenuKeyboard())
}

func (r *Router) handleHelp(ctx context.Context, message messageContext) error {
	return r.reply(ctx, message.ChatID, helpText(), r.mainMenuKeyboard())
}

func (r *Router) confirmUnlink(ctx context.Context, chatID int64, messageID int) error {
	return r.editOrSend(ctx, chatID, messageID,
		"<b>Отвязать Telegram?</b>\n\nДоступ к финансовым данным через бота будет закрыт. Данные в веб-приложении сохранятся.",
		ports.NewKeyboard(
			ports.ButtonRow(
				ports.CallbackButton("Да, отвязать", "set:unlink:yes"),
				ports.CallbackButton("Отмена", "set"),
			),
		))
}
