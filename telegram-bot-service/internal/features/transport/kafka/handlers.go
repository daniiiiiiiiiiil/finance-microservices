package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	corekafka "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/kafka"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/metrics"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/logger"
	"go.uber.org/zap"
)

type Handlers struct {
	notifications ports.NotificationServiceInterface
	dedup         ports.DedupStore
	logger        *logger.Logger
	dedupTTL      time.Duration
}

func NewHandlers(
	notifications ports.NotificationServiceInterface,
	dedup ports.DedupStore,
	logger *logger.Logger,
	dedupTTL time.Duration,
) *Handlers {
	return &Handlers{
		notifications: notifications,
		dedup:         dedup,
		logger:        logger,
		dedupTTL:      dedupTTL,
	}
}

func (h *Handlers) Register(consumer *corekafka.Consumer) error {
	handlers := map[string]corekafka.EventHandler{
		corekafka.EventTypeBudgetExceeded:     h.handleBudgetExceeded,
		corekafka.EventTypePlanCreated:        h.handlePlanCreated,
		corekafka.EventTypeTransactionCreated: h.handleTransaction,
		corekafka.EventTypeUserDeleted:        h.handleUserDeleted,
	}
	for eventType, handler := range handlers {
		if err := consumer.RegisterHandler(eventType, handler); err != nil {
			return fmt.Errorf("register handler %s: %w", eventType, err)
		}
	}
	return nil
}

func (h *Handlers) handleBudgetExceeded(ctx context.Context, event corekafka.Event) error {
	var payload corekafka.BudgetExceededEvent
	if err := json.Unmarshal(event.Data, &payload); err != nil {
		return fmt.Errorf("parse budget exceeded event: %w", err)
	}
	if payload.UserID <= 0 {
		return fmt.Errorf("budget exceeded event: invalid user id")
	}
	return h.process(ctx, event, func() error {
		return h.notifications.HandleBudgetExceeded(ctx, payload)
	})
}

func (h *Handlers) handlePlanCreated(ctx context.Context, event corekafka.Event) error {
	var payload corekafka.PlanCreatedEvent
	if err := json.Unmarshal(event.Data, &payload); err != nil {
		return fmt.Errorf("parse plan created event: %w", err)
	}
	if payload.UserID <= 0 {
		return fmt.Errorf("plan created event: invalid user id")
	}
	return h.process(ctx, event, func() error {
		return h.notifications.HandlePlanCreated(ctx, payload)
	})
}

func (h *Handlers) handleTransaction(ctx context.Context, event corekafka.Event) error {
	var payload corekafka.TransactionEvent
	if err := json.Unmarshal(event.Data, &payload); err != nil {
		return fmt.Errorf("parse transaction event: %w", err)
	}
	if payload.UserID <= 0 {
		return fmt.Errorf("transaction event: invalid user id")
	}
	return h.process(ctx, event, func() error {
		return h.notifications.HandleTransaction(ctx, payload)
	})
}

func (h *Handlers) handleUserDeleted(ctx context.Context, event corekafka.Event) error {
	var payload corekafka.UserDeletedEvent
	if err := json.Unmarshal(event.Data, &payload); err != nil {
		return fmt.Errorf("parse user deleted event: %w", err)
	}
	if payload.UserID <= 0 {
		return fmt.Errorf("user deleted event: invalid user id")
	}
	return h.process(ctx, event, func() error {
		return h.notifications.HandleUserDeleted(ctx, payload)
	})
}

func (h *Handlers) process(ctx context.Context, event corekafka.Event, action func() error) error {
	key := fmt.Sprintf("telegram:event:%s:%s", event.Type, event.ID)
	claimed := true
	if event.ID != "" {
		var err error
		claimed, err = h.dedup.Claim(ctx, key, int(h.dedupTTL.Seconds()))
		if err != nil {
			h.logger.Warn("event dedup claim failed", zap.String("event_type", event.Type), zap.Error(err))
		} else if !claimed {
			h.logger.Debug("duplicate event ignored", zap.String("event_type", event.Type))
			return nil
		}
	}

	if err := action(); err != nil {
		if claimed && event.ID != "" {
			_ = h.dedup.Release(ctx, key)
		}
		metrics.RecordKafkaMessage(event.Type, false)
		return err
	}
	metrics.RecordKafkaMessage(event.Type, true)
	return nil
}
