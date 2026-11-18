package service

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/kafka"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/ports"
	apperrors "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/errors"
)

type fakeTelegramAPI struct {
	mu         sync.Mutex
	messages   []string
	attempts   int
	failTimes  int
	alwaysFail error
	failWith   error
	documents  int
	photos     int
	locations  int
	callbacks  int
	messageID  int
}

func (f *fakeTelegramAPI) SendMessage(_ context.Context, _ int64, text string, _ *ports.Keyboard) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.attempts++
	if f.alwaysFail != nil {
		return 0, f.alwaysFail
	}
	if f.attempts <= f.failTimes {
		if f.failWith != nil {
			return 0, f.failWith
		}
		return 0, errors.New("temporary telegram error")
	}
	f.messages = append(f.messages, text)
	f.messageID++
	return f.messageID, nil
}

func (f *fakeTelegramAPI) EditMessage(context.Context, int64, int, string, *ports.Keyboard) error {
	return nil
}

func (f *fakeTelegramAPI) EditReplyMarkup(context.Context, int64, int, *ports.Keyboard) error {
	return nil
}

func (f *fakeTelegramAPI) DeleteMessage(context.Context, int64, int) error { return nil }

func (f *fakeTelegramAPI) AnswerCallback(context.Context, string, string, bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.callbacks++
	return nil
}

func (f *fakeTelegramAPI) SendPhoto(context.Context, int64, []byte, string, string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.photos++
	return nil
}

func (f *fakeTelegramAPI) SendDocument(context.Context, int64, []byte, string, string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.documents++
	return nil
}

func (f *fakeTelegramAPI) SendLocation(context.Context, int64, float64, float64) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.locations++
	return nil
}

func (f *fakeTelegramAPI) DownloadFile(context.Context, string) ([]byte, error) {
	return []byte("image"), nil
}

func (f *fakeTelegramAPI) SendChatAction(context.Context, int64, string) error { return nil }

func (f *fakeTelegramAPI) SetWebhook(context.Context, string, string) error { return nil }

func (f *fakeTelegramAPI) GetMe(context.Context) (ports.BotIdentity, error) {
	return ports.BotIdentity{ID: 1, Username: "test_bot"}, nil
}

func (f *fakeTelegramAPI) StartPolling(context.Context, func(context.Context, []byte) error) error {
	return nil
}

func (f *fakeTelegramAPI) sentMessages() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.messages...)
}

func TestNotificationBudgetExceededDelivered(t *testing.T) {
	accounts := newFakeAccountRepository()
	settings := newFakeNotificationRepository()
	if err := accounts.CreateAccount(context.Background(), domain.TelegramAccount{TelegramUserID: 100, ChatID: 200, UserID: 5}); err != nil {
		t.Fatalf("seed account: %v", err)
	}
	defaults := domain.DefaultNotificationSettings(5, time.Now())
	if err := settings.UpsertSettings(context.Background(), defaults); err != nil {
		t.Fatalf("seed settings: %v", err)
	}
	telegram := &fakeTelegramAPI{}
	notifier := NewNotificationService(accounts, settings, telegram, newTestLogger(t), 3)

	err := notifier.HandleBudgetExceeded(context.Background(), kafka.BudgetExceededEvent{
		BudgetID:    1,
		UserID:      5,
		Category:    "Продукты",
		AmountLimit: 10000,
		Spent:       11250,
		ExceededBy:  1250,
		Period:      "month",
	})
	if err != nil {
		t.Fatalf("handle budget exceeded: %v", err)
	}
	messages := telegram.sentMessages()
	if len(messages) != 1 {
		t.Fatalf("expected one message, got %d", len(messages))
	}
	if !strings.Contains(messages[0], "Продукты") || !strings.Contains(messages[0], "Превышен бюджет") {
		t.Fatalf("unexpected message: %s", messages[0])
	}
}

func TestNotificationDisabledIsSkipped(t *testing.T) {
	accounts := newFakeAccountRepository()
	settings := newFakeNotificationRepository()
	if err := accounts.CreateAccount(context.Background(), domain.TelegramAccount{TelegramUserID: 100, ChatID: 200, UserID: 5}); err != nil {
		t.Fatalf("seed account: %v", err)
	}
	config := domain.DefaultNotificationSettings(5, time.Now())
	config.BudgetAlerts = false
	if err := settings.UpsertSettings(context.Background(), config); err != nil {
		t.Fatalf("seed settings: %v", err)
	}
	telegram := &fakeTelegramAPI{}
	notifier := NewNotificationService(accounts, settings, telegram, newTestLogger(t), 3)

	if err := notifier.HandleBudgetExceeded(context.Background(), kafka.BudgetExceededEvent{UserID: 5, Category: "Кафе"}); err != nil {
		t.Fatalf("handle budget exceeded: %v", err)
	}
	if len(telegram.sentMessages()) != 0 {
		t.Fatal("expected no message when alerts are disabled")
	}
}

func TestNotificationRetriesTransientErrors(t *testing.T) {
	accounts := newFakeAccountRepository()
	settings := newFakeNotificationRepository()
	if err := accounts.CreateAccount(context.Background(), domain.TelegramAccount{TelegramUserID: 100, ChatID: 200, UserID: 5}); err != nil {
		t.Fatalf("seed account: %v", err)
	}
	if err := settings.UpsertSettings(context.Background(), domain.DefaultNotificationSettings(5, time.Now())); err != nil {
		t.Fatalf("seed settings: %v", err)
	}
	telegram := &fakeTelegramAPI{failTimes: 2}
	notifier := NewNotificationService(accounts, settings, telegram, newTestLogger(t), 3)
	notifier.wait = func(context.Context, time.Duration) error { return nil }

	if err := notifier.HandlePlanCreated(context.Background(), kafka.PlanCreatedEvent{UserID: 5, TotalCost: 1000, Mode: "self_drive"}); err != nil {
		t.Fatalf("handle plan created: %v", err)
	}
	if telegram.attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", telegram.attempts)
	}
	if len(telegram.sentMessages()) != 1 {
		t.Fatal("expected message after retries")
	}
}

func TestNotificationForbiddenIsNotRetried(t *testing.T) {
	accounts := newFakeAccountRepository()
	settings := newFakeNotificationRepository()
	if err := accounts.CreateAccount(context.Background(), domain.TelegramAccount{TelegramUserID: 100, ChatID: 200, UserID: 5}); err != nil {
		t.Fatalf("seed account: %v", err)
	}
	if err := settings.UpsertSettings(context.Background(), domain.DefaultNotificationSettings(5, time.Now())); err != nil {
		t.Fatalf("seed settings: %v", err)
	}
	telegram := &fakeTelegramAPI{alwaysFail: apperrors.ErrTelegramForbidden}
	notifier := NewNotificationService(accounts, settings, telegram, newTestLogger(t), 3)
	notifier.wait = func(context.Context, time.Duration) error { return nil }

	if err := notifier.HandlePlanCreated(context.Background(), kafka.PlanCreatedEvent{UserID: 5}); err != nil {
		t.Fatalf("expected forbidden to be swallowed, got %v", err)
	}
	if telegram.attempts != 1 {
		t.Fatalf("expected single attempt, got %d", telegram.attempts)
	}
}

func TestNotificationForUnlinkedUserIsSkipped(t *testing.T) {
	notifier := NewNotificationService(newFakeAccountRepository(), newFakeNotificationRepository(), &fakeTelegramAPI{}, newTestLogger(t), 3)
	if err := notifier.HandleTransaction(context.Background(), kafka.TransactionEvent{UserID: 77, TypeTransaction: "expense", Amount: 10}); err != nil {
		t.Fatalf("expected no error for unlinked user, got %v", err)
	}
}

func TestNotificationUserDeletedCleansUp(t *testing.T) {
	accounts := newFakeAccountRepository()
	settings := newFakeNotificationRepository()
	if err := accounts.CreateAccount(context.Background(), domain.TelegramAccount{TelegramUserID: 100, ChatID: 200, UserID: 5}); err != nil {
		t.Fatalf("seed account: %v", err)
	}
	if err := settings.UpsertSettings(context.Background(), domain.DefaultNotificationSettings(5, time.Now())); err != nil {
		t.Fatalf("seed settings: %v", err)
	}
	notifier := NewNotificationService(accounts, settings, &fakeTelegramAPI{}, newTestLogger(t), 1)
	if err := notifier.HandleUserDeleted(context.Background(), kafka.UserDeletedEvent{UserID: 5}); err != nil {
		t.Fatalf("handle user deleted: %v", err)
	}
	if _, err := accounts.GetAccountByUserID(context.Background(), 5); err == nil {
		t.Fatal("expected account to be deleted")
	}
	if _, ok := settings.settings[5]; ok {
		t.Fatal("expected settings to be deleted")
	}
}

func TestNotificationDeliveryFailureReturnsError(t *testing.T) {
	accounts := newFakeAccountRepository()
	settings := newFakeNotificationRepository()
	if err := accounts.CreateAccount(context.Background(), domain.TelegramAccount{TelegramUserID: 100, ChatID: 200, UserID: 5}); err != nil {
		t.Fatalf("seed account: %v", err)
	}
	if err := settings.UpsertSettings(context.Background(), domain.DefaultNotificationSettings(5, time.Now())); err != nil {
		t.Fatalf("seed settings: %v", err)
	}
	telegram := &fakeTelegramAPI{alwaysFail: errors.New("telegram unavailable")}
	notifier := NewNotificationService(accounts, settings, telegram, newTestLogger(t), 2)
	notifier.wait = func(context.Context, time.Duration) error { return nil }

	err := notifier.HandlePlanCreated(context.Background(), kafka.PlanCreatedEvent{UserID: 5})
	if err == nil {
		t.Fatal("expected delivery error")
	}
	if telegram.attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", telegram.attempts)
	}
}
