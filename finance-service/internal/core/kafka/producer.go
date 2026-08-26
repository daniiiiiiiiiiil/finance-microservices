package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/pkg/logger"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Producer struct {
	writer *kafka.Writer
	config Config
	logger logger.Logger
}

func NewProducer(config Config, logger logger.Logger) *Producer {
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

func (p *Producer) SendEvent(ctx context.Context, eventType string, data interface{}) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal event data error: %w", err)
	}

	event := Event{
		ID:        generateEventID(),
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      dataBytes,
	}
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event error: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(eventType),
		Value: eventBytes,
		Time:  time.Now(),
		Headers: []kafka.Header{
			{
				Key:   "event_type",
				Value: []byte(eventType),
			},
			{
				Key:   "timestamp",
				Value: []byte(event.Timestamp.Format(time.RFC3339)),
			},
		},
	}
	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		p.logger.Error("write messages error", zap.Error(err))
		return fmt.Errorf("write messages error: %w", err)
	}
	p.logger.Debug("send event", zap.String("event_type", eventType))
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

func generateEventID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Nanosecond())
}
