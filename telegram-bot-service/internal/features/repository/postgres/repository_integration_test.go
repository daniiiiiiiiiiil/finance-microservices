package postgres

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/repository/postgres/pool/pgx"
	apperrors "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/errors"
)

func integrationPool(t *testing.T) *pgx.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	connectionPool, err := pgx.NewPoolWithDSN(ctx, dsn, 5*time.Second, 5)
	if err != nil {
		t.Fatalf("connect to postgres: %v", err)
	}
	t.Cleanup(connectionPool.Close)

	migration, err := os.ReadFile("../../../../migrations/000001_init.up.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	for _, statement := range strings.Split(string(migration), ";") {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}
		if _, err := connectionPool.Exec(ctx, statement); err != nil {
			t.Fatalf("apply migration statement: %v", err)
		}
	}
	return connectionPool
}

func cleanupIntegrationData(t *testing.T, connectionPool *pgx.Pool) {
	t.Helper()
	ctx := context.Background()
	for _, statement := range []string{
		"DELETE FROM telegram.notification_settings",
		"DELETE FROM telegram.link_codes",
		"DELETE FROM telegram.accounts",
	} {
		if _, err := connectionPool.Exec(ctx, statement); err != nil {
			t.Fatalf("cleanup: %v", err)
		}
	}
}

func TestIntegrationLinkCodeLifecycle(t *testing.T) {
	connectionPool := integrationPool(t)
	cleanupIntegrationData(t, connectionPool)
	t.Cleanup(func() { cleanupIntegrationData(t, connectionPool) })
	repository := NewRepository(connectionPool)
	ctx := context.Background()

	code := domain.LinkCode{
		UserID:    42,
		CodeHash:  "hash-integration",
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	if err := repository.CreateLinkCode(ctx, code); err != nil {
		t.Fatalf("create link code: %v", err)
	}
	stored, err := repository.GetLinkCodeByHash(ctx, code.CodeHash)
	if err != nil {
		t.Fatalf("get link code: %v", err)
	}
	if stored.UserID != 42 || stored.Used() {
		t.Fatalf("unexpected link code: %+v", stored)
	}
	if err := repository.MarkLinkCodeUsed(ctx, stored.ID); err != nil {
		t.Fatalf("mark used: %v", err)
	}
	used, err := repository.GetLinkCodeByHash(ctx, code.CodeHash)
	if err != nil {
		t.Fatalf("get used code: %v", err)
	}
	if !used.Used() {
		t.Fatal("expected code to be used")
	}
	if err := repository.MarkLinkCodeUsed(ctx, stored.ID); !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("expected conflict on double use, got %v", err)
	}
	if _, err := repository.GetLinkCodeByHash(ctx, "missing"); !errors.Is(err, apperrors.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestIntegrationAccountLifecycle(t *testing.T) {
	connectionPool := integrationPool(t)
	cleanupIntegrationData(t, connectionPool)
	t.Cleanup(func() { cleanupIntegrationData(t, connectionPool) })
	repository := NewRepository(connectionPool)
	ctx := context.Background()

	username := "integration_user"
	account := domain.TelegramAccount{
		TelegramUserID: 900001,
		ChatID:         900001,
		UserID:         42,
		Username:       &username,
	}
	if err := repository.CreateAccount(ctx, account); err != nil {
		t.Fatalf("create account: %v", err)
	}
	byTelegram, err := repository.GetAccountByTelegramID(ctx, account.TelegramUserID)
	if err != nil {
		t.Fatalf("get by telegram: %v", err)
	}
	if byTelegram.UserID != 42 || byTelegram.Username == nil || *byTelegram.Username != username {
		t.Fatalf("unexpected account: %+v", byTelegram)
	}
	byUser, err := repository.GetAccountByUserID(ctx, 42)
	if err != nil {
		t.Fatalf("get by user: %v", err)
	}
	if byUser.TelegramUserID != account.TelegramUserID {
		t.Fatalf("unexpected account: %+v", byUser)
	}
	if err := repository.CreateAccount(ctx, domain.TelegramAccount{TelegramUserID: 900002, ChatID: 900002, UserID: 42}); !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("expected conflict for duplicate user, got %v", err)
	}
	if err := repository.DeleteAccountByUserID(ctx, 42); err != nil {
		t.Fatalf("delete account: %v", err)
	}
	if _, err := repository.GetAccountByUserID(ctx, 42); !errors.Is(err, apperrors.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestIntegrationNotificationSettings(t *testing.T) {
	connectionPool := integrationPool(t)
	cleanupIntegrationData(t, connectionPool)
	t.Cleanup(func() { cleanupIntegrationData(t, connectionPool) })
	repository := NewRepository(connectionPool)
	ctx := context.Background()

	settings := domain.DefaultNotificationSettings(42, time.Now())
	settings.FinanceAlerts = true
	settings.BudgetAlerts = false
	if err := repository.UpsertSettings(ctx, settings); err != nil {
		t.Fatalf("upsert settings: %v", err)
	}
	stored, err := repository.GetSettings(ctx, 42)
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	if stored.FinanceAlerts != true || stored.BudgetAlerts != false {
		t.Fatalf("unexpected settings: %+v", stored)
	}
	settings.PlanAlerts = false
	if err := repository.UpsertSettings(ctx, settings); err != nil {
		t.Fatalf("update settings: %v", err)
	}
	stored, err = repository.GetSettings(ctx, 42)
	if err != nil {
		t.Fatalf("get updated settings: %v", err)
	}
	if stored.PlanAlerts != false {
		t.Fatalf("expected plan alerts disabled, got %+v", stored)
	}
	if err := repository.DeleteSettings(ctx, 42); err != nil {
		t.Fatalf("delete settings: %v", err)
	}
	if _, err := repository.GetSettings(ctx, 42); !errors.Is(err, apperrors.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}
