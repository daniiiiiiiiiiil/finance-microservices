package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/cache"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/domain"
	apperrors "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/errors"
	goredis "github.com/redis/go-redis/v9"
)

type DialogStore struct {
	client *cache.RedisClient
	ttl    time.Duration
}

func NewDialogStore(client *cache.RedisClient, ttl time.Duration) *DialogStore {
	return &DialogStore{client: client, ttl: ttl}
}

func (s *DialogStore) Get(ctx context.Context, chatID int64) (domain.DialogState, error) {
	key := dialogKey(chatID)
	var state domain.DialogState
	if err := s.client.Get(ctx, key, &state); err != nil {
		if errors.Is(err, goredis.Nil) {
			return domain.DialogState{}, fmt.Errorf("dialog state: %w", apperrors.ErrNotFound)
		}
		return domain.DialogState{}, fmt.Errorf("get dialog state: %w", err)
	}
	return state, nil
}

func (s *DialogStore) Set(ctx context.Context, state domain.DialogState) error {
	state.UpdatedAt = time.Now().UTC()
	if err := s.client.Set(ctx, dialogKey(state.ChatID), state, s.ttl); err != nil {
		return fmt.Errorf("set dialog state: %w", err)
	}
	return nil
}

func (s *DialogStore) Delete(ctx context.Context, chatID int64) error {
	if err := s.client.Delete(ctx, dialogKey(chatID)); err != nil {
		return fmt.Errorf("delete dialog state: %w", err)
	}
	return nil
}

func dialogKey(chatID int64) string {
	return fmt.Sprintf("telegram:dialog:%d", chatID)
}
