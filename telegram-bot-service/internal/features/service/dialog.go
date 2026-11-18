package service

import (
	"context"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/ports"
	apperrors "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/errors"
)

type DialogService struct {
	store ports.DialogStore
}

func NewDialogService(store ports.DialogStore) *DialogService {
	return &DialogService{store: store}
}

func (s *DialogService) Get(ctx context.Context, chatID int64) (domain.DialogState, bool) {
	state, err := s.store.Get(ctx, chatID)
	if err != nil {
		return domain.DialogState{}, false
	}
	return state, true
}

func (s *DialogService) Start(
	ctx context.Context,
	chatID int64,
	userID int,
	flow domain.Flow,
	step string,
	data map[string]string,
) (domain.DialogState, error) {
	if chatID == 0 || userID <= 0 {
		return domain.DialogState{}, fmt.Errorf("chat and user are required: %w", apperrors.ErrInvalidArgument)
	}
	state := domain.DialogState{
		ChatID: chatID,
		UserID: userID,
		Flow:   flow,
		Step:   step,
		Data:   copyData(data),
	}
	if err := s.store.Set(ctx, state); err != nil {
		return domain.DialogState{}, err
	}
	return state, nil
}

func (s *DialogService) Update(
	ctx context.Context,
	state domain.DialogState,
	step string,
	data map[string]string,
) (domain.DialogState, error) {
	state.Step = step
	merged := copyData(state.Data)
	for key, value := range data {
		merged[key] = value
	}
	state.Data = merged
	if err := s.store.Set(ctx, state); err != nil {
		return domain.DialogState{}, err
	}
	return state, nil
}

func (s *DialogService) Clear(ctx context.Context, chatID int64) error {
	return s.store.Delete(ctx, chatID)
}

func copyData(data map[string]string) map[string]string {
	result := make(map[string]string, len(data))
	for key, value := range data {
		result[key] = value
	}
	return result
}
