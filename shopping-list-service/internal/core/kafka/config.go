package kafka

import (
	"os"
	"time"
)

type Config struct {
	Brokers       []string
	Topic         string
	ConsumerGroup string
	MaxRetries    int
	RetryBackoff  time.Duration
}

func NewConfig() Config {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		topic = "finance-events"
	}

	consumerGroup := os.Getenv("KAFKA_CONSUMER_GROUP")
	if consumerGroup == "" {
		consumerGroup = "shopping-list-service"
	}

	return Config{
		Brokers:       []string{brokers},
		Topic:         topic,
		ConsumerGroup: consumerGroup,
		MaxRetries:    3,
		RetryBackoff:  time.Second,
	}
}

func (c Config) GetBrokers() []string {
	return c.Brokers
}
