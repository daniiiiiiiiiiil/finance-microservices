package kafka

import (
	"os"
	"strings"
)

const (
	EventTypeTransactionCreated = "transaction.created"
	EventTypeTransactionUpdated = "transaction.updated"
	EventTypeTransactionDeleted = "transaction.deleted"
	EventTypeBudgetCreated      = "budget.created"
	EventTypeBudgetUpdated      = "budget.updated"
	EventTypeBudgetDeleted      = "budget.deleted"
	EventTypeBudgetExceeded     = "budget.exceeded"
	EventTypePlanCreated        = "ai.plan_created"
	EventTypeUserDeleted        = "user.deleted"
)

type Config struct {
	Brokers       []string
	Topic         string
	ConsumerGroup string
	MaxRetries    int
}

func NewConfig() Config {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	parts := strings.Split(brokers, ",")
	normalized := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			normalized = append(normalized, part)
		}
	}
	if len(normalized) == 0 {
		normalized = []string{"localhost:9092"}
	}

	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		topic = "finance-events"
	}

	group := os.Getenv("KAFKA_CONSUMER_GROUP")
	if group == "" {
		group = "telegram-bot-service"
	}

	return Config{
		Brokers:       normalized,
		Topic:         topic,
		ConsumerGroup: group,
		MaxRetries:    3,
	}
}
