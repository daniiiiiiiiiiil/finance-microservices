package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/currency-service/pkg/logger"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Consumer struct {
	reader   *kafka.Reader
	config   Config
	logger   logger.Logger
	handlers map[string]func(ctx context.Context, event Event) error
}

func NewConsumer(config Config, logger logger.Logger) *Consumer {
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
		handlers: make(map[string]func(ctx context.Context, event Event) error),
	}
}

func (c *Consumer) RegisterHandler(eventType string, handler func(ctx context.Context, event Event) error) {
	c.handlers[eventType] = handler
	c.logger.Debug("register handler", zap.String("eventType", eventType))
}

func (c *Consumer) Start(ctx context.Context) error {
	c.logger.Info("start consumer", zap.String("group", c.config.ConsumerGroup), zap.String("topic", c.config.Topic))

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("stop consumer", zap.String("group", c.config.ConsumerGroup), zap.String("topic", c.config.Topic))
			return nil
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				c.logger.Error("error reading message", zap.String("group", c.config.ConsumerGroup), zap.String("topic", c.config.Topic))
				continue
			}
			if err := c.handleMessage(ctx, msg); err != nil {
				c.logger.Error("error handling message",
					zap.String("group", c.config.ConsumerGroup),
					zap.String("topic", c.config.Topic),
					zap.ByteString("message", msg.Value),
					zap.Int("partition", msg.Partition),
					zap.Int64("offset", msg.Offset))
				continue
			}
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				c.logger.Error("error committing message")
			}
		}
	}
}

func (c *Consumer) handleMessage(ctx context.Context, msg kafka.Message) error {
	var event Event
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("unmarshal event: %w", err)
	}
	handler, ok := c.handlers[event.Type]
	if !ok {
		c.logger.Warn("unknown event type", zap.String("eventType", event.Type))
		return nil
	}
	return handler(ctx, event)
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
