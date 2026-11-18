package redis

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/cache"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/domain"
)

func integrationRedis(t *testing.T) *cache.RedisClient {
	t.Helper()
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("TEST_REDIS_ADDR is not set")
	}
	client := cache.NewRedisClient(addr)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx); err != nil {
		t.Fatalf("ping redis: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func TestIntegrationDialogStore(t *testing.T) {
	client := integrationRedis(t)
	store := NewDialogStore(client, time.Minute)
	ctx := context.Background()
	chatID := time.Now().UnixNano()

	state := domain.DialogState{
		ChatID: chatID,
		UserID: 7,
		Flow:   domain.FlowTransaction,
		Step:   domain.StepTransactionAmount,
		Data:   map[string]string{"type": "expense"},
	}
	if err := store.Set(ctx, state); err != nil {
		t.Fatalf("set state: %v", err)
	}
	stored, err := store.Get(ctx, chatID)
	if err != nil {
		t.Fatalf("get state: %v", err)
	}
	if stored.UserID != 7 || stored.Get("type") != "expense" {
		t.Fatalf("unexpected state: %+v", stored)
	}
	if err := store.Delete(ctx, chatID); err != nil {
		t.Fatalf("delete state: %v", err)
	}
	if _, err := store.Get(ctx, chatID); err == nil {
		t.Fatal("expected state to be deleted")
	}
}

func TestIntegrationDedupStore(t *testing.T) {
	client := integrationRedis(t)
	store := NewDedupStore(client)
	ctx := context.Background()
	key := fmt.Sprintf("telegram:test:dedup:%d", time.Now().UnixNano())
	t.Cleanup(func() { _ = client.Delete(ctx, key) })

	claimed, err := store.Claim(ctx, key, 60)
	if err != nil || !claimed {
		t.Fatalf("first claim: claimed=%v err=%v", claimed, err)
	}
	claimed, err = store.Claim(ctx, key, 60)
	if err != nil || claimed {
		t.Fatalf("second claim: claimed=%v err=%v", claimed, err)
	}
	if err := store.Release(ctx, key); err != nil {
		t.Fatalf("release: %v", err)
	}
	claimed, err = store.Claim(ctx, key, 60)
	if err != nil || !claimed {
		t.Fatalf("claim after release: claimed=%v err=%v", claimed, err)
	}
}
