package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/domain"
	corekafka "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/kafka"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/logger"
)

type fakeNotificationService struct {
	budgetCalls      int
	planCalls        int
	transactionCalls int
	userDeletedCalls int
	err              error
}

func (f *fakeNotificationService) Send(context.Context, int, domain.Notification) error { return nil }

func (f *fakeNotificationService) HandleBudgetExceeded(context.Context, corekafka.BudgetExceededEvent) error {
	f.budgetCalls++
	return f.err
}

func (f *fakeNotificationService) HandlePlanCreated(context.Context, corekafka.PlanCreatedEvent) error {
	f.planCalls++
	return f.err
}

func (f *fakeNotificationService) HandleTransaction(context.Context, corekafka.TransactionEvent) error {
	f.transactionCalls++
	return f.err
}

func (f *fakeNotificationService) HandleUserDeleted(context.Context, corekafka.UserDeletedEvent) error {
	f.userDeletedCalls++
	return f.err
}

type fakeKafkaDedup struct {
	keys map[string]bool
}

func newFakeKafkaDedup() *fakeKafkaDedup {
	return &fakeKafkaDedup{keys: make(map[string]bool)}
}

func (f *fakeKafkaDedup) Claim(_ context.Context, key string, _ int) (bool, error) {
	if f.keys[key] {
		return false, nil
	}
	f.keys[key] = true
	return true, nil
}

func (f *fakeKafkaDedup) Release(_ context.Context, key string) error {
	delete(f.keys, key)
	return nil
}

func newKafkaTestLogger(t *testing.T) *logger.Logger {
	t.Helper()
	log, err := logger.NewLogger(logger.LoggerConfig{Level: "error", Folder: t.TempDir()})
	if err != nil {
		t.Fatalf("create logger: %v", err)
	}
	t.Cleanup(log.Close)
	return log
}

func budgetEvent(t *testing.T, id string, userID int) corekafka.Event {
	t.Helper()
	payload, err := json.Marshal(corekafka.BudgetExceededEvent{UserID: userID, Category: "Продукты", AmountLimit: 1000, Spent: 1200, ExceededBy: 200})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return corekafka.Event{ID: id, Type: corekafka.EventTypeBudgetExceeded, Timestamp: time.Now(), Data: payload}
}

func TestHandlersBudgetExceeded(t *testing.T) {
	notifications := &fakeNotificationService{}
	handlers := NewHandlers(notifications, newFakeKafkaDedup(), newKafkaTestLogger(t), time.Hour)

	if err := handlers.handleBudgetExceeded(context.Background(), budgetEvent(t, "event-1", 5)); err != nil {
		t.Fatalf("handle budget: %v", err)
	}
	if notifications.budgetCalls != 1 {
		t.Fatalf("expected 1 call, got %d", notifications.budgetCalls)
	}
}

func TestHandlersIgnoresDuplicateEvents(t *testing.T) {
	notifications := &fakeNotificationService{}
	handlers := NewHandlers(notifications, newFakeKafkaDedup(), newKafkaTestLogger(t), time.Hour)
	event := budgetEvent(t, "event-1", 5)

	if err := handlers.handleBudgetExceeded(context.Background(), event); err != nil {
		t.Fatalf("first event: %v", err)
	}
	if err := handlers.handleBudgetExceeded(context.Background(), event); err != nil {
		t.Fatalf("duplicate event: %v", err)
	}
	if notifications.budgetCalls != 1 {
		t.Fatalf("expected duplicate to be skipped, got %d calls", notifications.budgetCalls)
	}
}

func TestHandlersReleasesDedupOnFailure(t *testing.T) {
	notifications := &fakeNotificationService{err: errors.New("telegram unavailable")}
	dedup := newFakeKafkaDedup()
	handlers := NewHandlers(notifications, dedup, newKafkaTestLogger(t), time.Hour)

	if err := handlers.handleBudgetExceeded(context.Background(), budgetEvent(t, "event-1", 5)); err == nil {
		t.Fatal("expected delivery error")
	}
	claimed, err := dedup.Claim(context.Background(), "telegram:event:budget.exceeded:event-1", 60)
	if err != nil || !claimed {
		t.Fatalf("expected dedup to be released, claimed=%v err=%v", claimed, err)
	}
}

func TestHandlersRejectsInvalidPayload(t *testing.T) {
	handlers := NewHandlers(&fakeNotificationService{}, newFakeKafkaDedup(), newKafkaTestLogger(t), time.Hour)
	event := corekafka.Event{ID: "event-1", Type: corekafka.EventTypeBudgetExceeded, Data: json.RawMessage(`{`)}
	if err := handlers.handleBudgetExceeded(context.Background(), event); err == nil {
		t.Fatal("expected payload error")
	}
}

func TestHandlersRequiresUserID(t *testing.T) {
	handlers := NewHandlers(&fakeNotificationService{}, newFakeKafkaDedup(), newKafkaTestLogger(t), time.Hour)
	if err := handlers.handleBudgetExceeded(context.Background(), budgetEvent(t, "event-1", 0)); err == nil {
		t.Fatal("expected user id validation error")
	}
}

func TestHandlersPlanCreated(t *testing.T) {
	notifications := &fakeNotificationService{}
	handlers := NewHandlers(notifications, newFakeKafkaDedup(), newKafkaTestLogger(t), time.Hour)
	payload, _ := json.Marshal(corekafka.PlanCreatedEvent{UserID: 5, PlanID: 3, TotalCost: 1000})
	event := corekafka.Event{ID: "event-2", Type: corekafka.EventTypePlanCreated, Data: payload}
	if err := handlers.handlePlanCreated(context.Background(), event); err != nil {
		t.Fatalf("handle plan: %v", err)
	}
	if notifications.planCalls != 1 {
		t.Fatalf("expected 1 plan call, got %d", notifications.planCalls)
	}
}

func TestHandlersTransactionCreated(t *testing.T) {
	notifications := &fakeNotificationService{}
	handlers := NewHandlers(notifications, newFakeKafkaDedup(), newKafkaTestLogger(t), time.Hour)
	payload, _ := json.Marshal(corekafka.TransactionEvent{UserID: 5, Amount: 100, TypeTransaction: "expense"})
	event := corekafka.Event{ID: "event-3", Type: corekafka.EventTypeTransactionCreated, Data: payload}
	if err := handlers.handleTransaction(context.Background(), event); err != nil {
		t.Fatalf("handle transaction: %v", err)
	}
	if notifications.transactionCalls != 1 {
		t.Fatalf("expected 1 transaction call, got %d", notifications.transactionCalls)
	}
}

func TestHandlersUserDeleted(t *testing.T) {
	notifications := &fakeNotificationService{}
	handlers := NewHandlers(notifications, newFakeKafkaDedup(), newKafkaTestLogger(t), time.Hour)
	payload, _ := json.Marshal(corekafka.UserDeletedEvent{UserID: 5})
	event := corekafka.Event{ID: "event-4", Type: corekafka.EventTypeUserDeleted, Data: payload}
	if err := handlers.handleUserDeleted(context.Background(), event); err != nil {
		t.Fatalf("handle user deleted: %v", err)
	}
	if notifications.userDeletedCalls != 1 {
		t.Fatalf("expected 1 user deleted call, got %d", notifications.userDeletedCalls)
	}
}
