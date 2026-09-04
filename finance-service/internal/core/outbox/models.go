package outbox

import (
	"encoding/json"
	"time"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusProcessed Status = "processed"
	StatusFailed    Status = "failed"
)

type OutboxMessage struct {
	ID            int64           `json:"id"`
	AggregateID   int64           `json:"aggregate_id"`
	AggregateType string          `json:"aggregate_type"`
	EventType     string          `json:"event_type"`
	Payload       json.RawMessage `json:"payload"`
	Headers       json.RawMessage `json:"headers"`
	Status        Status          `json:"status"`
	RetryCount    int             `json:"retry_count"`
	MaxRetries    int             `json:"max_retries"`
	CreatedAt     time.Time       `json:"created_at"`
	ProcessedAt   *time.Time      `json:"processed_at"`
	LastError     *string         `json:"last_error"`
	Version       int64           `json:"version"`
}

func NewOutboxMessage(aggregateID int64, aggregateType, eventType string, payload interface{}) (*OutboxMessage, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &OutboxMessage{
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		EventType:     eventType,
		Payload:       payloadBytes,
		Headers:       json.RawMessage(`{}`),
		Status:        StatusPending,
		MaxRetries:    3,
	}, nil
}
