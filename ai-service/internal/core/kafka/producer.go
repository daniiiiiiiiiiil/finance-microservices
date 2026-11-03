package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/pkg/logger"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Producer struct {
	writer *kafka.Writer
	config Config
	logger *logger.Logger
}

func NewProducer(config Config, logger *logger.Logger) *Producer {
	write := &kafka.Writer{
		Addr:         kafka.TCP(config.Brokers...),
		Topic:        config.Topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		MaxAttempts:  config.MaxRetries,
		BatchSize:    100,
		BatchTimeout: 100 * time.Millisecond,
	}
	return &Producer{
		writer: write,
		config: config,
		logger: logger,
	}
}

func (p *Producer) SendEvent(ctx context.Context, eventType string, data any) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal event data: %w", err)
	}
	event := Event{
		ID:        uuid.NewString(),
		Type:      eventType,
		Timestamp: time.Now().UTC(),
		Data:      dataBytes,
	}
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	if err := p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(eventType),
		Value: eventBytes,
	}); err != nil {
		p.logger.Error("kafka write failed",
			zap.String("event_type", eventType),
			zap.Error(err))
		return fmt.Errorf("write message: %w", err)
	}
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

func generateEventID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Nanosecond())
}
