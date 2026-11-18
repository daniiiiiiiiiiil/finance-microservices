package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/logger"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type EventHandler func(ctx context.Context, event Event) error

type Consumer struct {
	reader   *kafka.Reader
	config   Config
	logger   *logger.Logger
	handlers map[string]EventHandler
}

func NewConsumer(config Config, logger *logger.Logger) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        config.Brokers,
		GroupID:        config.ConsumerGroup,
		Topic:          config.Topic,
		MinBytes:       10e3,
		MaxBytes:       10e6,
		MaxWait:        1 * time.Second,
		CommitInterval: time.Second,
		StartOffset:    kafka.LastOffset,
	})
	return &Consumer{
		reader:   reader,
		config:   config,
		logger:   logger,
		handlers: make(map[string]EventHandler),
	}
}

func (c *Consumer) RegisterHandler(eventType string, handler EventHandler) error {
	if eventType == "" {
		return fmt.Errorf("event type is required")
	}
	if handler == nil {
		return fmt.Errorf("handler is required")
	}
	if _, exists := c.handlers[eventType]; exists {
		return fmt.Errorf("handler already registered for %s", eventType)
	}
	c.handlers[eventType] = handler
	return nil
}

func (c *Consumer) Start(ctx context.Context) error {
	c.logger.Info("kafka consumer started",
		zap.String("group", c.config.ConsumerGroup),
		zap.String("topic", c.config.Topic),
	)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("kafka consumer stopped",
				zap.String("group", c.config.ConsumerGroup),
				zap.String("topic", c.config.Topic),
			)
			return nil
		default:
		}

		message, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			c.logger.Error("kafka fetch error",
				zap.String("group", c.config.ConsumerGroup),
				zap.Error(err),
			)
			time.Sleep(time.Second)
			continue
		}

		if err := c.handleMessage(ctx, message); err != nil {
			c.logger.Error("kafka message handling error",
				zap.String("group", c.config.ConsumerGroup),
				zap.Int("partition", message.Partition),
				zap.Int64("offset", message.Offset),
				zap.Error(err),
			)
			continue
		}

		if err := c.reader.CommitMessages(ctx, message); err != nil {
			c.logger.Error("kafka commit error", zap.Error(err))
		}
	}
}

func (c *Consumer) handleMessage(ctx context.Context, message kafka.Message) error {
	var event Event
	if err := json.Unmarshal(message.Value, &event); err != nil {
		return fmt.Errorf("unmarshal event: %w", err)
	}
	if event.Type == "" {
		return fmt.Errorf("event type is empty")
	}
	handler, ok := c.handlers[event.Type]
	if !ok {
		c.logger.Debug("kafka event ignored", zap.String("event_type", event.Type))
		return nil
	}
	return handler(ctx, event)
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
