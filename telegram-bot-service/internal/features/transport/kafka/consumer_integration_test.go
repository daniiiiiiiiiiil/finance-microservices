package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	corekafka "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/kafka"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/logger"
	"github.com/segmentio/kafka-go"
)

func TestIntegrationKafkaConsumer(t *testing.T) {
	brokersEnv := os.Getenv("TEST_KAFKA_BROKERS")
	if brokersEnv == "" {
		t.Skip("TEST_KAFKA_BROKERS is not set")
	}
	brokers := strings.Split(brokersEnv, ",")
	for index := range brokers {
		brokers[index] = strings.TrimSpace(brokers[index])
	}
	topic := fmt.Sprintf("telegram-bot-integration-%d", time.Now().UnixNano())
	group := fmt.Sprintf("telegram-bot-integration-group-%d", time.Now().UnixNano())

	log, err := logger.NewLogger(logger.LoggerConfig{Level: "error", Folder: t.TempDir()})
	if err != nil {
		t.Fatalf("create logger: %v", err)
	}
	defer log.Close()

	consumer := corekafka.NewConsumer(corekafka.Config{
		Brokers:       brokers,
		Topic:         topic,
		ConsumerGroup: group,
		MaxRetries:    1,
	}, log)

	received := make(chan corekafka.Event, 1)
	if err := consumer.RegisterHandler("integration.test", func(_ context.Context, event corekafka.Event) error {
		received <- event
		return nil
	}); err != nil {
		t.Fatalf("register handler: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() {
		_ = consumer.Start(ctx)
		close(done)
	}()
	defer func() {
		cancel()
		_ = consumer.Close()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
		}
	}()

	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		MaxAttempts:  3,
	}
	defer writer.Close()

	time.Sleep(2 * time.Second)
	payload := []byte(`{"item":"milk"}`)
	event := corekafka.Event{ID: "integration-1", Type: "integration.test", Timestamp: time.Now(), Data: payload}
	encoded, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}
	if err := writer.WriteMessages(ctx, kafka.Message{Key: []byte("integration"), Value: encoded}); err != nil {
		t.Fatalf("write message: %v", err)
	}

	select {
	case event := <-received:
		if event.ID != "integration-1" || event.Type != "integration.test" {
			t.Fatalf("unexpected event: %+v", event)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("timed out waiting for kafka event")
	}
}
