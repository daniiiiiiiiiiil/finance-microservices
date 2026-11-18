package service

import (
	"context"
	"testing"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/domain"
	apperrors "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/errors"
)

type fakeDialogStore struct {
	states map[int64]domain.DialogState
}

func newFakeDialogStore() *fakeDialogStore {
	return &fakeDialogStore{states: make(map[int64]domain.DialogState)}
}

func (s *fakeDialogStore) Get(_ context.Context, chatID int64) (domain.DialogState, error) {
	state, ok := s.states[chatID]
	if !ok {
		return domain.DialogState{}, apperrors.ErrNotFound
	}
	return state, nil
}

func (s *fakeDialogStore) Set(_ context.Context, state domain.DialogState) error {
	state.UpdatedAt = time.Now().UTC()
	s.states[state.ChatID] = state
	return nil
}

func (s *fakeDialogStore) Delete(_ context.Context, chatID int64) error {
	delete(s.states, chatID)
	return nil
}

func TestDialogStartGetUpdateClear(t *testing.T) {
	store := newFakeDialogStore()
	service := NewDialogService(store)
	ctx := context.Background()

	if _, ok := service.Get(ctx, 10); ok {
		t.Fatal("expected no active dialog")
	}
	state, err := service.Start(ctx, 10, 1, domain.FlowTransaction, domain.StepTransactionAmount, map[string]string{"type": "expense"})
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if state.Get("type") != "expense" {
		t.Fatalf("unexpected state: %+v", state)
	}
	updated, err := service.Update(ctx, state, domain.StepTransactionCategory, map[string]string{"amount": "100"})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Step != domain.StepTransactionCategory || updated.Get("amount") != "100" || updated.Get("type") != "expense" {
		t.Fatalf("unexpected updated state: %+v", updated)
	}
	got, ok := service.Get(ctx, 10)
	if !ok || got.UserID != 1 {
		t.Fatalf("unexpected stored state: %+v", got)
	}
	if err := service.Clear(ctx, 10); err != nil {
		t.Fatalf("clear: %v", err)
	}
	if _, ok := service.Get(ctx, 10); ok {
		t.Fatal("expected dialog to be cleared")
	}
}

func TestDialogStartRejectsInvalidArguments(t *testing.T) {
	service := NewDialogService(newFakeDialogStore())
	if _, err := service.Start(context.Background(), 0, 0, domain.FlowBudget, domain.StepBudgetAmount, nil); err == nil {
		t.Fatal("expected validation error")
	}
}
