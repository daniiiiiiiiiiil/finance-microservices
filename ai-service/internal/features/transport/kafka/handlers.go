package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	corekafka "github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/core/kafka"
	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/pkg/logger"
	"go.uber.org/zap"
)

// Handlers — Kafka-хендлеры. Зависит от интерфейса, не от *shopping.Service.
type Handlers struct {
	service ports.PlanServiceInterface
	logger  *logger.Logger
}

func NewHandlers(svc ports.PlanServiceInterface, log *logger.Logger) *Handlers {
	return &Handlers{service: svc, logger: log}
}

func (h *Handlers) Register(consumer *Consumer) error {
	if err := consumer.RegisterHandler(corekafka.EventTypeUserDeleted, h.handleUserDeleted); err != nil {
		return fmt.Errorf("register %s: %w", corekafka.EventTypeUserDeleted, err)
	}
	return nil
}

func (h *Handlers) handleUserDeleted(ctx context.Context, event corekafka.Event) error {
	var payload corekafka.UserDeletedEvent
	if err := json.Unmarshal(event.Data, &payload); err != nil {
		return fmt.Errorf("parse user.deleted: %w", err)
	}

	h.logger.Info("received user.deleted", zap.Int("user_id", payload.UserID))

	if err := h.service.DeletePlansByUser(ctx, payload.UserID); err != nil {
		return fmt.Errorf("delete plans for user %d: %w", payload.UserID, err)
	}
	return nil
}
