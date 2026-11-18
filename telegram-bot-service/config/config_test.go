package config

import (
	"strings"
	"testing"
)

func clearTelegramConfigEnv(t *testing.T) {
	t.Helper()
	for _, name := range []string{
		"JWT_SECRET", "JWT_DURATION", "TELEGRAM_BOT_TOKEN", "TELEGRAM_MODE",
		"TELEGRAM_WEBHOOK_URL", "TELEGRAM_WEBHOOK_SECRET", "TELEGRAM_PUBLIC_URL",
		"TELEGRAM_WEBAPP_PATH", "TELEGRAM_WEBAPP_MAX_AGE", "TELEGRAM_WEBAPP_TOKEN_TTL",
		"TELEGRAM_SERVICE_TOKEN_TTL", "TELEGRAM_LINK_CODE_TTL", "TELEGRAM_DIALOG_TTL",
		"TELEGRAM_UPDATE_DEDUP_TTL", "TELEGRAM_EVENT_DEDUP_TTL", "TELEGRAM_HTTP_TIMEOUT",
		"TELEGRAM_NOTIFICATION_RETRIES", "TELEGRAM_HTTP_ADDR", "TELEGRAM_GRPC_ADDR",
		"TELEGRAM_METRICS_ADDR", "GATEWAY_URL", "GATEWAY_PUBLIC_URL", "REDIS_ADDR",
	} {
		t.Setenv(name, "")
	}
}

func TestNewConfigPollingMode(t *testing.T) {
	clearTelegramConfigEnv(t)
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("TELEGRAM_BOT_TOKEN", "token")
	t.Setenv("TELEGRAM_MODE", "polling")

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("new config: %v", err)
	}
	if cfg.TelegramMode != ModePolling {
		t.Fatalf("unexpected mode: %s", cfg.TelegramMode)
	}
	if cfg.TelegramHTTPAddr != ":8082" || cfg.TelegramGRPCAddr != ":50062" {
		t.Fatalf("unexpected addresses: %+v", cfg)
	}
	if cfg.WebAppTokenTTL.Minutes() != 15 {
		t.Fatalf("unexpected web app ttl: %v", cfg.WebAppTokenTTL)
	}
}

func TestNewConfigDerivesWebhookURL(t *testing.T) {
	clearTelegramConfigEnv(t)
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("TELEGRAM_BOT_TOKEN", "token")
	t.Setenv("TELEGRAM_MODE", "webhook")
	t.Setenv("TELEGRAM_WEBHOOK_SECRET", "webhook-secret")
	t.Setenv("TELEGRAM_PUBLIC_URL", "https://bot.example.com/")

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("new config: %v", err)
	}
	if cfg.TelegramWebhookURL != "https://bot.example.com/telegram/webhook" {
		t.Fatalf("unexpected webhook url: %s", cfg.TelegramWebhookURL)
	}
	if cfg.TelegramPublicURL != "https://bot.example.com" {
		t.Fatalf("unexpected public url: %s", cfg.TelegramPublicURL)
	}
	if cfg.WebAppURL() != "https://bot.example.com/telegram/webapp" {
		t.Fatalf("unexpected web app url: %s", cfg.WebAppURL())
	}
}

func TestNewConfigWebhookRequiresSecret(t *testing.T) {
	clearTelegramConfigEnv(t)
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("TELEGRAM_BOT_TOKEN", "token")
	t.Setenv("TELEGRAM_MODE", "webhook")
	_, err := NewConfig()
	if err == nil || !strings.Contains(err.Error(), "TELEGRAM_WEBHOOK_SECRET") {
		t.Fatalf("expected webhook secret error, got %v", err)
	}
}

func TestNewConfigRejectsInvalidMode(t *testing.T) {
	clearTelegramConfigEnv(t)
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("TELEGRAM_BOT_TOKEN", "token")
	t.Setenv("TELEGRAM_MODE", "invalid")
	if _, err := NewConfig(); err == nil {
		t.Fatal("expected invalid mode error")
	}
}

func TestNewConfigRequiresToken(t *testing.T) {
	clearTelegramConfigEnv(t)
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("TELEGRAM_MODE", "polling")
	if _, err := NewConfig(); err == nil || !strings.Contains(err.Error(), "TELEGRAM_BOT_TOKEN") {
		t.Fatalf("expected token error, got %v", err)
	}
}

func TestNewConfigRejectsInvalidWebhookSecret(t *testing.T) {
	clearTelegramConfigEnv(t)
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("TELEGRAM_BOT_TOKEN", "token")
	t.Setenv("TELEGRAM_MODE", "webhook")
	t.Setenv("TELEGRAM_WEBHOOK_SECRET", "bad secret with spaces")
	if _, err := NewConfig(); err == nil {
		t.Fatal("expected webhook secret validation error")
	}
}
