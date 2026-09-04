package service

import (
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/pkg/logger"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type OutboxPublisher struct {
	repo     ports.OutboxRepositoryInterface
	producer ports.KafkaProducerInterface
	logger   *logger.Logger
}

func NewOutboxPublisher(
	repo ports.OutboxRepositoryInterface,
	producer ports.KafkaProducerInterface,
	logger *logger.Logger,
) *OutboxPublisher {
	return &OutboxPublisher{
		repo:     repo,
		producer: producer,
		logger:   logger,
	}
}

func (p *OutboxPublisher) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				p.logger.Info("OutboxPublisher shutting down")
				return
			case <-ticker.C:
				p.publishPending(ctx)
			}
		}
	}()
}

func (p *OutboxPublisher) publishPending(ctx context.Context) {
	events, err := p.repo.GetPending(ctx, 100)
	if err != nil {
		p.logger.Error("failed to get pending events", zap.Error(err))
		return
	}

	for _, event := range events {
		if err := p.producer.SendEvent(ctx, event.EventType, event.EventPayload); err != nil {
			p.logger.Error("failed to send event",
				zap.String("event_type", event.EventType),
				zap.String("event_id", event.ID),
				zap.Error(err))

			if markErr := p.repo.MarkFailed(ctx, event.ID, err.Error()); markErr != nil {
				p.logger.Error("failed to mark failed event", zap.Error(markErr))
			}
			continue
		}

		if err := p.repo.MarkProcessed(ctx, event.ID); err != nil {
			p.logger.Error("failed to mark processed event", zap.Error(err))
			continue
		}

		p.logger.Debug("outbox event sent",
			zap.String("event_type", event.EventType),
			zap.String("event_id", event.ID),
		)
	}
}
