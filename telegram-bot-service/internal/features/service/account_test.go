package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/domain"
	apperrors "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/errors"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/logger"
)

type fakeAccountRepository struct {
	linkCodes      map[string]domain.LinkCode
	byTelegram     map[int64]domain.TelegramAccount
	byUser         map[int]domain.TelegramAccount
	nextLinkCodeID int64
}

func newFakeAccountRepository() *fakeAccountRepository {
	return &fakeAccountRepository{
		linkCodes:  make(map[string]domain.LinkCode),
		byTelegram: make(map[int64]domain.TelegramAccount),
		byUser:     make(map[int]domain.TelegramAccount),
	}
}

func (r *fakeAccountRepository) CreateLinkCode(_ context.Context, code domain.LinkCode) error {
	r.nextLinkCodeID++
	code.ID = r.nextLinkCodeID
	code.CreatedAt = time.Now().UTC()
	r.linkCodes[code.CodeHash] = code
	return nil
}

func (r *fakeAccountRepository) GetLinkCodeByHash(_ context.Context, hash string) (domain.LinkCode, error) {
	code, ok := r.linkCodes[hash]
	if !ok {
		return domain.LinkCode{}, apperrors.ErrNotFound
	}
	return code, nil
}

func (r *fakeAccountRepository) MarkLinkCodeUsed(_ context.Context, id int64) error {
	for hash, code := range r.linkCodes {
		if code.ID != id {
			continue
		}
		if code.UsedAt != nil {
			return apperrors.ErrConflict
		}
		now := time.Now().UTC()
		code.UsedAt = &now
		r.linkCodes[hash] = code
		return nil
	}
	return apperrors.ErrNotFound
}

func (r *fakeAccountRepository) IncrementLinkCodeAttempts(_ context.Context, id int64) error {
	for hash, code := range r.linkCodes {
		if code.ID == id {
			code.Attempts++
			r.linkCodes[hash] = code
		}
	}
	return nil
}

func (r *fakeAccountRepository) CreateAccount(_ context.Context, account domain.TelegramAccount) error {
	if _, exists := r.byTelegram[account.TelegramUserID]; exists {
		return apperrors.ErrConflict
	}
	if _, exists := r.byUser[account.UserID]; exists {
		return apperrors.ErrConflict
	}
	account.CreatedAt = time.Now().UTC()
	account.UpdatedAt = account.CreatedAt
	r.byTelegram[account.TelegramUserID] = account
	r.byUser[account.UserID] = account
	return nil
}

func (r *fakeAccountRepository) GetAccountByTelegramID(_ context.Context, telegramUserID int64) (domain.TelegramAccount, error) {
	account, ok := r.byTelegram[telegramUserID]
	if !ok {
		return domain.TelegramAccount{}, apperrors.ErrNotFound
	}
	return account, nil
}

func (r *fakeAccountRepository) GetAccountByUserID(_ context.Context, userID int) (domain.TelegramAccount, error) {
	account, ok := r.byUser[userID]
	if !ok {
		return domain.TelegramAccount{}, apperrors.ErrNotFound
	}
	return account, nil
}

func (r *fakeAccountRepository) DeleteAccountByTelegramID(_ context.Context, telegramUserID int64) error {
	account, ok := r.byTelegram[telegramUserID]
	if !ok {
		return apperrors.ErrNotFound
	}
	delete(r.byTelegram, telegramUserID)
	delete(r.byUser, account.UserID)
	return nil
}

func (r *fakeAccountRepository) DeleteAccountByUserID(_ context.Context, userID int) error {
	account, ok := r.byUser[userID]
	if !ok {
		return apperrors.ErrNotFound
	}
	delete(r.byUser, userID)
	delete(r.byTelegram, account.TelegramUserID)
	return nil
}

type fakeNotificationRepository struct {
	settings map[int]domain.NotificationSettings
}

func newFakeNotificationRepository() *fakeNotificationRepository {
	return &fakeNotificationRepository{settings: make(map[int]domain.NotificationSettings)}
}

func (r *fakeNotificationRepository) GetSettings(_ context.Context, userID int) (domain.NotificationSettings, error) {
	settings, ok := r.settings[userID]
	if !ok {
		return domain.NotificationSettings{}, apperrors.ErrNotFound
	}
	return settings, nil
}

func (r *fakeNotificationRepository) UpsertSettings(_ context.Context, settings domain.NotificationSettings) error {
	r.settings[settings.UserID] = settings
	return nil
}

func (r *fakeNotificationRepository) DeleteSettings(_ context.Context, userID int) error {
	delete(r.settings, userID)
	return nil
}

func newTestLogger(t *testing.T) *logger.Logger {
	t.Helper()
	log, err := logger.NewLogger(logger.LoggerConfig{Level: "error", Folder: t.TempDir()})
	if err != nil {
		t.Fatalf("create logger: %v", err)
	}
	t.Cleanup(log.Close)
	return log
}

func TestCreateLinkCodeStoresHashAndExpiry(t *testing.T) {
	accounts := newFakeAccountRepository()
	settings := newFakeNotificationRepository()
	service := NewAccountService(accounts, settings, newTestLogger(t), 10*time.Minute)

	result, err := service.CreateLinkCode(context.Background(), 42)
	if err != nil {
		t.Fatalf("create link code: %v", err)
	}
	if len(result.Code) != linkCodeLength {
		t.Fatalf("expected code length %d, got %d", linkCodeLength, len(result.Code))
	}
	if result.ExpiresAt.Before(time.Now()) {
		t.Fatal("expected future expiry")
	}
	code, err := accounts.GetLinkCodeByHash(context.Background(), hashLinkCode(result.Code))
	if err != nil {
		t.Fatalf("get stored code: %v", err)
	}
	if code.UserID != 42 {
		t.Fatalf("expected user 42, got %d", code.UserID)
	}
	if strings.Contains(code.CodeHash, result.Code) {
		t.Fatal("link code must not be stored in plain text")
	}
}

func TestConsumeLinkCodeLinksAccountOnce(t *testing.T) {
	accounts := newFakeAccountRepository()
	settings := newFakeNotificationRepository()
	service := NewAccountService(accounts, settings, newTestLogger(t), 10*time.Minute)

	result, err := service.CreateLinkCode(context.Background(), 7)
	if err != nil {
		t.Fatalf("create link code: %v", err)
	}
	username := "alice"
	account, err := service.ConsumeLinkCode(context.Background(), result.Code, 1000, 2000, &username)
	if err != nil {
		t.Fatalf("consume link code: %v", err)
	}
	if account.UserID != 7 || account.TelegramUserID != 1000 || account.ChatID != 2000 {
		t.Fatalf("unexpected account: %+v", account)
	}
	if account.Username == nil || *account.Username != username {
		t.Fatal("expected username to be stored")
	}
	if _, ok := settings.settings[7]; !ok {
		t.Fatal("expected default notification settings")
	}
	if _, err := service.ConsumeLinkCode(context.Background(), result.Code, 1000, 2000, &username); !errors.Is(err, ErrCodeUsed) {
		t.Fatalf("expected ErrCodeUsed, got %v", err)
	}
}

func TestConsumeLinkCodeRejectsExpiredCode(t *testing.T) {
	accounts := newFakeAccountRepository()
	settings := newFakeNotificationRepository()
	service := NewAccountService(accounts, settings, newTestLogger(t), 10*time.Minute)
	base := time.Now().UTC()
	service.now = func() time.Time { return base }

	result, err := service.CreateLinkCode(context.Background(), 7)
	if err != nil {
		t.Fatalf("create link code: %v", err)
	}
	service.now = func() time.Time { return base.Add(11 * time.Minute) }
	if _, err := service.ConsumeLinkCode(context.Background(), result.Code, 1000, 2000, nil); !errors.Is(err, ErrCodeInvalid) {
		t.Fatalf("expected ErrCodeInvalid, got %v", err)
	}
}

func TestConsumeLinkCodeRejectsTelegramBoundToAnotherUser(t *testing.T) {
	accounts := newFakeAccountRepository()
	settings := newFakeNotificationRepository()
	service := NewAccountService(accounts, settings, newTestLogger(t), 10*time.Minute)

	if err := accounts.CreateAccount(context.Background(), domain.TelegramAccount{TelegramUserID: 1000, ChatID: 1000, UserID: 1}); err != nil {
		t.Fatalf("seed account: %v", err)
	}
	result, err := service.CreateLinkCode(context.Background(), 2)
	if err != nil {
		t.Fatalf("create link code: %v", err)
	}
	if _, err := service.ConsumeLinkCode(context.Background(), result.Code, 1000, 1000, nil); !errors.Is(err, ErrAlreadyLinked) {
		t.Fatalf("expected ErrAlreadyLinked, got %v", err)
	}
}

func TestConsumeLinkCodeRejectsUserWithAnotherTelegram(t *testing.T) {
	accounts := newFakeAccountRepository()
	settings := newFakeNotificationRepository()
	service := NewAccountService(accounts, settings, newTestLogger(t), 10*time.Minute)

	result, err := service.CreateLinkCode(context.Background(), 5)
	if err != nil {
		t.Fatalf("create link code: %v", err)
	}
	if err := accounts.CreateAccount(context.Background(), domain.TelegramAccount{TelegramUserID: 2000, ChatID: 2000, UserID: 5}); err != nil {
		t.Fatalf("seed account: %v", err)
	}
	if _, err := service.ConsumeLinkCode(context.Background(), result.Code, 3000, 3000, nil); !errors.Is(err, ErrAlreadyLinked) {
		t.Fatalf("expected ErrAlreadyLinked, got %v", err)
	}
}

func TestUnlinkRemovesAccountAndSettings(t *testing.T) {
	accounts := newFakeAccountRepository()
	settings := newFakeNotificationRepository()
	service := NewAccountService(accounts, settings, newTestLogger(t), 10*time.Minute)

	if err := accounts.CreateAccount(context.Background(), domain.TelegramAccount{TelegramUserID: 1000, ChatID: 1000, UserID: 9}); err != nil {
		t.Fatalf("seed account: %v", err)
	}
	if err := settings.UpsertSettings(context.Background(), domain.DefaultNotificationSettings(9, time.Now())); err != nil {
		t.Fatalf("seed settings: %v", err)
	}
	if err := service.UnlinkByTelegramID(context.Background(), 1000); err != nil {
		t.Fatalf("unlink: %v", err)
	}
	if _, err := service.Resolve(context.Background(), 1000); !errors.Is(err, ErrNotLinked) {
		t.Fatalf("expected ErrNotLinked, got %v", err)
	}
	if _, ok := settings.settings[9]; ok {
		t.Fatal("expected notification settings to be deleted")
	}
}

func TestGetSettingsReturnsDefaultsForNewUser(t *testing.T) {
	service := NewAccountService(newFakeAccountRepository(), newFakeNotificationRepository(), newTestLogger(t), 10*time.Minute)
	settings, err := service.GetSettings(context.Background(), 15)
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	if !settings.BudgetAlerts || !settings.PlanAlerts || settings.FinanceAlerts {
		t.Fatalf("unexpected defaults: %+v", settings)
	}
}

func TestUpdateSettingsPersistsValues(t *testing.T) {
	settingsRepo := newFakeNotificationRepository()
	service := NewAccountService(newFakeAccountRepository(), settingsRepo, newTestLogger(t), 10*time.Minute)
	settings := domain.DefaultNotificationSettings(3, time.Now())
	settings.FinanceAlerts = true
	settings.BudgetAlerts = false
	if err := service.UpdateSettings(context.Background(), settings); err != nil {
		t.Fatalf("update settings: %v", err)
	}
	stored, err := settingsRepo.GetSettings(context.Background(), 3)
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	if stored.FinanceAlerts != true || stored.BudgetAlerts != false {
		t.Fatalf("settings not persisted: %+v", stored)
	}
}
