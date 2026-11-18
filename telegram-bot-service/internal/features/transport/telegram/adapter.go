package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/metrics"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/ports"
	apperrors "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/errors"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/logger"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"
)

const maxDownloadBytes = 10 << 20

type Adapter struct {
	client *bot.Bot
	logger *logger.Logger
}

func NewAdapter(token string, logger *logger.Logger) (*Adapter, error) {
	client, err := bot.New(token, bot.WithErrorsHandler(func(err error) {
		logger.Error("telegram bot error", zap.Error(err))
	}))
	if err != nil {
		return nil, fmt.Errorf("create telegram bot: %w", err)
	}
	return &Adapter{client: client, logger: logger}, nil
}

func (a *Adapter) SendMessage(ctx context.Context, chatID int64, text string, keyboard *ports.Keyboard) (int, error) {
	message, err := a.client.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:             chatID,
		Text:               text,
		ParseMode:          models.ParseModeHTML,
		ReplyMarkup:        markup(keyboard),
		LinkPreviewOptions: &models.LinkPreviewOptions{IsDisabled: boolPointer(true)},
	})
	if err != nil {
		metrics.RecordTelegramAPI("sendMessage", false)
		return 0, mapError(err)
	}
	metrics.RecordTelegramAPI("sendMessage", true)
	return message.ID, nil
}

func (a *Adapter) EditMessage(ctx context.Context, chatID int64, messageID int, text string, keyboard *ports.Keyboard) error {
	if messageID <= 0 {
		_, err := a.SendMessage(ctx, chatID, text, keyboard)
		return err
	}
	_, err := a.client.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:             chatID,
		MessageID:          messageID,
		Text:               text,
		ParseMode:          models.ParseModeHTML,
		ReplyMarkup:        markup(keyboard),
		LinkPreviewOptions: &models.LinkPreviewOptions{IsDisabled: boolPointer(true)},
	})
	if err != nil {
		if isNotModified(err) {
			return nil
		}
		metrics.RecordTelegramAPI("editMessageText", false)
		return mapError(err)
	}
	metrics.RecordTelegramAPI("editMessageText", true)
	return nil
}

func (a *Adapter) EditReplyMarkup(ctx context.Context, chatID int64, messageID int, keyboard *ports.Keyboard) error {
	if messageID <= 0 {
		return nil
	}
	_, err := a.client.EditMessageReplyMarkup(ctx, &bot.EditMessageReplyMarkupParams{
		ChatID:      chatID,
		MessageID:   messageID,
		ReplyMarkup: markup(keyboard),
	})
	if err != nil {
		if isNotModified(err) {
			return nil
		}
		metrics.RecordTelegramAPI("editMessageReplyMarkup", false)
		return mapError(err)
	}
	metrics.RecordTelegramAPI("editMessageReplyMarkup", true)
	return nil
}

func (a *Adapter) DeleteMessage(ctx context.Context, chatID int64, messageID int) error {
	if messageID <= 0 {
		return nil
	}
	_, err := a.client.DeleteMessage(ctx, &bot.DeleteMessageParams{ChatID: chatID, MessageID: messageID})
	if err != nil {
		if isMessageNotFound(err) {
			return nil
		}
		metrics.RecordTelegramAPI("deleteMessage", false)
		return mapError(err)
	}
	metrics.RecordTelegramAPI("deleteMessage", true)
	return nil
}

func (a *Adapter) AnswerCallback(ctx context.Context, callbackID, text string, alert bool) error {
	if callbackID == "" {
		return nil
	}
	_, err := a.client.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: callbackID,
		Text:            text,
		ShowAlert:       alert,
	})
	if err != nil {
		metrics.RecordTelegramAPI("answerCallbackQuery", false)
		return mapError(err)
	}
	metrics.RecordTelegramAPI("answerCallbackQuery", true)
	return nil
}

func (a *Adapter) SendPhoto(ctx context.Context, chatID int64, data []byte, filename, caption string) error {
	_, err := a.client.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID:    chatID,
		Photo:     &models.InputFileUpload{Filename: filename, Data: bytes.NewReader(data)},
		Caption:   caption,
		ParseMode: models.ParseModeHTML,
	})
	if err != nil {
		metrics.RecordTelegramAPI("sendPhoto", false)
		return mapError(err)
	}
	metrics.RecordTelegramAPI("sendPhoto", true)
	return nil
}

func (a *Adapter) SendDocument(ctx context.Context, chatID int64, data []byte, filename, caption string) error {
	_, err := a.client.SendDocument(ctx, &bot.SendDocumentParams{
		ChatID:    chatID,
		Document:  &models.InputFileUpload{Filename: filename, Data: bytes.NewReader(data)},
		Caption:   caption,
		ParseMode: models.ParseModeHTML,
	})
	if err != nil {
		metrics.RecordTelegramAPI("sendDocument", false)
		return mapError(err)
	}
	metrics.RecordTelegramAPI("sendDocument", true)
	return nil
}

func (a *Adapter) SendLocation(ctx context.Context, chatID int64, latitude, longitude float64) error {
	_, err := a.client.SendLocation(ctx, &bot.SendLocationParams{
		ChatID:    chatID,
		Latitude:  latitude,
		Longitude: longitude,
	})
	if err != nil {
		metrics.RecordTelegramAPI("sendLocation", false)
		return mapError(err)
	}
	metrics.RecordTelegramAPI("sendLocation", true)
	return nil
}

func (a *Adapter) DownloadFile(ctx context.Context, fileID string) ([]byte, error) {
	if fileID == "" {
		return nil, fmt.Errorf("file id is required")
	}
	file, err := a.client.GetFile(ctx, &bot.GetFileParams{FileID: fileID})
	if err != nil {
		metrics.RecordTelegramAPI("getFile", false)
		return nil, mapError(err)
	}
	metrics.RecordTelegramAPI("getFile", true)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, a.client.FileDownloadLink(file), nil)
	if err != nil {
		return nil, fmt.Errorf("create file request: %w", err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		metrics.RecordTelegramAPI("downloadFile", false)
		return nil, fmt.Errorf("download telegram file: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		metrics.RecordTelegramAPI("downloadFile", false)
		return nil, fmt.Errorf("download telegram file: status %d", response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxDownloadBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read telegram file: %w", err)
	}
	if len(data) > maxDownloadBytes {
		return nil, fmt.Errorf("telegram file exceeds %d bytes", maxDownloadBytes)
	}
	metrics.RecordTelegramAPI("downloadFile", true)
	return data, nil
}

func (a *Adapter) SendChatAction(ctx context.Context, chatID int64, action string) error {
	chatAction := models.ChatAction(action)
	if chatAction == "" {
		chatAction = models.ChatActionTyping
	}
	_, err := a.client.SendChatAction(ctx, &bot.SendChatActionParams{ChatID: chatID, Action: chatAction})
	if err != nil {
		metrics.RecordTelegramAPI("sendChatAction", false)
		return mapError(err)
	}
	metrics.RecordTelegramAPI("sendChatAction", true)
	return nil
}

func (a *Adapter) SetWebhook(ctx context.Context, url, secretToken string) error {
	_, err := a.client.SetWebhook(ctx, &bot.SetWebhookParams{
		URL:            url,
		SecretToken:    secretToken,
		AllowedUpdates: []string{"message", "callback_query"},
	})
	if err != nil {
		return fmt.Errorf("set webhook: %w", mapError(err))
	}
	return nil
}

func (a *Adapter) GetMe(ctx context.Context) (ports.BotIdentity, error) {
	user, err := a.client.GetMe(ctx)
	if err != nil {
		return ports.BotIdentity{}, mapError(err)
	}
	return ports.BotIdentity{ID: user.ID, Username: user.Username}, nil
}

func (a *Adapter) StartPolling(ctx context.Context, handler func(ctx context.Context, update []byte) error) error {
	if handler == nil {
		return fmt.Errorf("polling handler is required")
	}
	a.client.RegisterHandlerMatchFunc(
		func(update *models.Update) bool { return update != nil },
		func(ctx context.Context, _ *bot.Bot, update *models.Update) {
			payload, err := json.Marshal(update)
			if err != nil {
				a.logger.Error("marshal telegram update", zap.Error(err))
				return
			}
			if err := handler(ctx, payload); err != nil {
				a.logger.Error("process telegram update", zap.Error(err))
			}
		},
	)
	a.client.Start(ctx)
	return nil
}

func markup(keyboard *ports.Keyboard) models.ReplyMarkup {
	if keyboard == nil || len(keyboard.Rows) == 0 {
		return nil
	}
	rows := make([][]models.InlineKeyboardButton, 0, len(keyboard.Rows))
	for _, row := range keyboard.Rows {
		if len(row) == 0 {
			continue
		}
		buttons := make([]models.InlineKeyboardButton, 0, len(row))
		for _, button := range row {
			item := models.InlineKeyboardButton{Text: button.Text}
			switch {
			case button.WebAppURL != "":
				item.WebApp = &models.WebAppInfo{URL: button.WebAppURL}
			case button.URL != "":
				item.URL = button.URL
			default:
				item.CallbackData = button.CallbackData
			}
			buttons = append(buttons, item)
		}
		rows = append(rows, buttons)
	}
	if len(rows) == 0 {
		return nil
	}
	return &models.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, bot.ErrorForbidden) {
		return fmt.Errorf("%v: %w", err, apperrors.ErrTelegramForbidden)
	}
	if bot.IsTooManyRequestsError(err) {
		return fmt.Errorf("%v: %w", err, apperrors.ErrUnavailable)
	}
	return fmt.Errorf("telegram api: %w", err)
}

func isNotModified(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "message is not modified")
}

func isMessageNotFound(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "message to delete not found") ||
		strings.Contains(message, "message can't be deleted")
}

func boolPointer(value bool) *bool {
	return &value
}
