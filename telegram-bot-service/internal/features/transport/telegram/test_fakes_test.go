package telegram

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/features/service"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/logger"
	"github.com/go-telegram/bot/models"
)

type fakeAccountService struct {
	mu           sync.Mutex
	linked       bool
	account      domain.TelegramAccount
	consumedCode string
	settings     domain.NotificationSettings
	unlinkCalled bool
}

func (f *fakeAccountService) CreateLinkCode(context.Context, int) (ports.LinkCodeResult, error) {
	return ports.LinkCodeResult{Code: "ABCDEFGH", ExpiresAt: time.Now().Add(time.Minute)}, nil
}

func (f *fakeAccountService) ConsumeLinkCode(_ context.Context, code string, telegramUserID, chatID int64, username *string) (domain.TelegramAccount, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.consumedCode = code
	f.linked = true
	f.account = domain.TelegramAccount{TelegramUserID: telegramUserID, ChatID: chatID, UserID: 7, Username: username, CreatedAt: time.Now()}
	return f.account, nil
}

func (f *fakeAccountService) GetLink(context.Context, int) (domain.TelegramAccount, error) {
	if !f.linked {
		return domain.TelegramAccount{}, service.ErrNotLinked
	}
	return f.account, nil
}

func (f *fakeAccountService) Resolve(context.Context, int64) (domain.TelegramAccount, error) {
	if !f.linked {
		return domain.TelegramAccount{}, service.ErrNotLinked
	}
	return f.account, nil
}

func (f *fakeAccountService) UnlinkByTelegramID(context.Context, int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.unlinkCalled = true
	f.linked = false
	return nil
}

func (f *fakeAccountService) UnlinkByUserID(context.Context, int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.unlinkCalled = true
	f.linked = false
	return nil
}

func (f *fakeAccountService) GetSettings(context.Context, int) (domain.NotificationSettings, error) {
	if f.settings.UserID == 0 {
		return domain.DefaultNotificationSettings(7, time.Now()), nil
	}
	return f.settings, nil
}

func (f *fakeAccountService) UpdateSettings(_ context.Context, settings domain.NotificationSettings) error {
	f.settings = settings
	return nil
}

type fakeDialogService struct {
	mu     sync.Mutex
	states map[int64]domain.DialogState
}

func newFakeDialogService() *fakeDialogService {
	return &fakeDialogService{states: make(map[int64]domain.DialogState)}
}

func (f *fakeDialogService) Get(_ context.Context, chatID int64) (domain.DialogState, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	state, ok := f.states[chatID]
	return state, ok
}

func (f *fakeDialogService) Start(_ context.Context, chatID int64, userID int, flow domain.Flow, step string, data map[string]string) (domain.DialogState, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	state := domain.DialogState{ChatID: chatID, UserID: userID, Flow: flow, Step: step, Data: copyTestData(data)}
	f.states[chatID] = state
	return state, nil
}

func (f *fakeDialogService) Update(_ context.Context, state domain.DialogState, step string, data map[string]string) (domain.DialogState, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	state.Step = step
	merged := copyTestData(state.Data)
	for key, value := range data {
		merged[key] = value
	}
	state.Data = merged
	f.states[state.ChatID] = state
	return state, nil
}

func (f *fakeDialogService) Clear(_ context.Context, chatID int64) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.states, chatID)
	return nil
}

func copyTestData(data map[string]string) map[string]string {
	result := make(map[string]string, len(data))
	for key, value := range data {
		result[key] = value
	}
	return result
}

type createdTransaction struct {
	UserID      int
	Type        string
	Amount      float64
	Category    string
	Description string
	ID          int
}

type fakePlatform struct {
	ports.PlatformClient
	mu              sync.Mutex
	categories      []string
	created         *createdTransaction
	updated         *createdTransaction
	createdShopping *domain.ShoppingInput
	updatedShopping *domain.ShoppingInput
	shoppingItem    *domain.ShoppingItem
	dashboard       domain.Dashboard
}

func (f *fakePlatform) GetCategories(context.Context, int) ([]string, error) {
	if len(f.categories) == 0 {
		return []string{"Продукты", "Кафе"}, nil
	}
	return f.categories, nil
}

func (f *fakePlatform) GetDashboard(context.Context, int) (domain.Dashboard, error) {
	return f.dashboard, nil
}

func (f *fakePlatform) GetTransaction(_ context.Context, _ int, id int) (domain.Transaction, error) {
	return domain.Transaction{ID: id, Type: "expense", Amount: 100, Category: "Кафе", Description: "обед"}, nil
}

func (f *fakePlatform) DeleteTransaction(context.Context, int, int) error {
	return nil
}

func (f *fakePlatform) CreateTransaction(_ context.Context, userID int, transactionType string, amount float64, category string, description string) (domain.Transaction, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.created = &createdTransaction{UserID: userID, Type: transactionType, Amount: amount, Category: category, Description: description, ID: 55}
	return domain.Transaction{ID: 55, Type: transactionType, Amount: amount, Category: category, Description: description}, nil
}

func (f *fakePlatform) UpdateTransaction(_ context.Context, userID, id int, transactionType string, amount float64, category string, description string) (domain.Transaction, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.updated = &createdTransaction{UserID: userID, Type: transactionType, Amount: amount, Category: category, Description: description, ID: id}
	return domain.Transaction{ID: id, Type: transactionType, Amount: amount, Category: category, Description: description}, nil
}

func (f *fakePlatform) CreateShopping(_ context.Context, _ int, input domain.ShoppingInput) (domain.ShoppingItem, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.createdShopping = &input
	return domain.ShoppingItem{ID: 11, Title: input.Title, AmountNow: input.AmountNow, AmountFinish: input.AmountFinish}, nil
}

func (f *fakePlatform) GetShopping(_ context.Context, _ int, id int) (domain.ShoppingItem, error) {
	if f.shoppingItem != nil {
		return *f.shoppingItem, nil
	}
	return domain.ShoppingItem{ID: id, Title: "Список", AmountNow: 100, AmountFinish: 1000}, nil
}

func (f *fakePlatform) UpdateShopping(_ context.Context, _ int, item domain.ShoppingItem, input domain.ShoppingInput) (domain.ShoppingItem, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.updatedShopping = &input
	item.Title = input.Title
	item.AmountNow = input.AmountNow
	item.AmountFinish = input.AmountFinish
	imageKey := "shopping/photo.jpg"
	item.ImageKey = &imageKey
	return item, nil
}

type capturedMessage struct {
	ChatID   int64
	Text     string
	Keyboard *ports.Keyboard
	Edited   bool
}

type fakeTelegram struct {
	mu        sync.Mutex
	messages  []capturedMessage
	callbacks int
	messageCh chan capturedMessage
	nextID    int
}

func newFakeTelegram() *fakeTelegram {
	return &fakeTelegram{messageCh: make(chan capturedMessage, 32)}
}

func (f *fakeTelegram) record(message capturedMessage) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.messages = append(f.messages, message)
	select {
	case f.messageCh <- message:
	default:
	}
}

func (f *fakeTelegram) SendMessage(_ context.Context, chatID int64, text string, keyboard *ports.Keyboard) (int, error) {
	f.record(capturedMessage{ChatID: chatID, Text: text, Keyboard: keyboard})
	f.mu.Lock()
	defer f.mu.Unlock()
	f.nextID++
	return f.nextID, nil
}

func (f *fakeTelegram) EditMessage(_ context.Context, chatID int64, _ int, text string, keyboard *ports.Keyboard) error {
	f.record(capturedMessage{ChatID: chatID, Text: text, Keyboard: keyboard, Edited: true})
	return nil
}

func (f *fakeTelegram) EditReplyMarkup(context.Context, int64, int, *ports.Keyboard) error {
	return nil
}

func (f *fakeTelegram) DeleteMessage(context.Context, int64, int) error { return nil }

func (f *fakeTelegram) AnswerCallback(context.Context, string, string, bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.callbacks++
	return nil
}

func (f *fakeTelegram) SendPhoto(context.Context, int64, []byte, string, string) error { return nil }

func (f *fakeTelegram) SendDocument(context.Context, int64, []byte, string, string) error { return nil }

func (f *fakeTelegram) SendLocation(context.Context, int64, float64, float64) error { return nil }

func (f *fakeTelegram) DownloadFile(context.Context, string) ([]byte, error) {
	return []byte("image"), nil
}

func (f *fakeTelegram) SendChatAction(context.Context, int64, string) error { return nil }

func (f *fakeTelegram) SetWebhook(context.Context, string, string) error { return nil }

func (f *fakeTelegram) GetMe(context.Context) (ports.BotIdentity, error) {
	return ports.BotIdentity{ID: 1, Username: "test_bot"}, nil
}

func (f *fakeTelegram) StartPolling(context.Context, func(context.Context, []byte) error) error {
	return nil
}

func (f *fakeTelegram) lastMessage() (capturedMessage, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.messages) == 0 {
		return capturedMessage{}, false
	}
	return f.messages[len(f.messages)-1], true
}

func (f *fakeTelegram) allTexts() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	texts := make([]string, 0, len(f.messages))
	for _, message := range f.messages {
		texts = append(texts, message.Text)
	}
	return texts
}

type fakeDedup struct {
	mu   sync.Mutex
	keys map[string]bool
}

func newFakeDedup() *fakeDedup {
	return &fakeDedup{keys: make(map[string]bool)}
}

func (f *fakeDedup) Claim(_ context.Context, key string, _ int) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.keys[key] {
		return false, nil
	}
	f.keys[key] = true
	return true, nil
}

func (f *fakeDedup) Release(_ context.Context, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.keys, key)
	return nil
}

func newRouterTestLogger(t *testing.T) *logger.Logger {
	t.Helper()
	log, err := logger.NewLogger(logger.LoggerConfig{Level: "error", Folder: t.TempDir()})
	if err != nil {
		t.Fatalf("create logger: %v", err)
	}
	t.Cleanup(log.Close)
	return log
}

func newTestRouter(t *testing.T, accounts *fakeAccountService, dialogs *fakeDialogService, platform *fakePlatform, telegram *fakeTelegram) *Router {
	t.Helper()
	return NewRouter(telegram, accounts, dialogs, platform, newRouterTestLogger(t), "https://example.com/telegram/webapp")
}

func textUpdate(id int64, telegramUserID, chatID int64, text string) models.Update {
	return models.Update{
		ID: id,
		Message: &models.Message{
			ID:   1,
			From: &models.User{ID: telegramUserID, FirstName: "Alice"},
			Chat: models.Chat{ID: chatID, Type: models.ChatTypePrivate},
			Text: text,
		},
	}
}

func callbackUpdate(id int64, telegramUserID, chatID int64, data string) models.Update {
	return models.Update{
		ID: id,
		CallbackQuery: &models.CallbackQuery{
			ID:   "callback-1",
			From: models.User{ID: telegramUserID, FirstName: "Alice"},
			Data: data,
			Message: models.MaybeInaccessibleMessage{
				Message: &models.Message{ID: 2, Chat: models.Chat{ID: chatID, Type: models.ChatTypePrivate}},
			},
		},
	}
}

func photoUpdate(id int64, telegramUserID, chatID int64, fileID string) models.Update {
	return models.Update{
		ID: id,
		Message: &models.Message{
			ID:   3,
			From: &models.User{ID: telegramUserID, FirstName: "Alice"},
			Chat: models.Chat{ID: chatID, Type: models.ChatTypePrivate},
			Photo: []models.PhotoSize{
				{FileID: "small", Width: 100, Height: 100},
				{FileID: fileID, Width: 800, Height: 600},
			},
		},
	}
}

func containsText(messages []string, fragment string) bool {
	for _, message := range messages {
		if strings.Contains(message, fragment) {
			return true
		}
	}
	return false
}
