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
	return Config{
		Brokers:      []string{brokers},
		Topic:        "finance-events",
		MaxRetries:   3,
		RetryBackoff: time.Second}
}

func (c Config) GetBrokers() []string {
	return c.Brokers
}
