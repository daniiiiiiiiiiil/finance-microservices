package domain

import (
	"encoding/json"
	"time"
)

type OutboxEvent struct {
	ID            string          `json:"id"`
	AggregateID   int64           `json:"aggregate_id"`
	AggregateType string          `json:"aggregate_type"`
	EventType     string          `json:"event_type"`
	EventPayload  json.RawMessage `json:"payload"`
	CreatedAt     time.Time       `json:"created_at"`
	ProcessedAt   time.Time       `json:"processed_at"`
	Status        string          `json:"status"`
	ErrorMessage  string          `json:"last_error"`
	RetryCount    int             `json:"retry_count"`
}

func NewOutboxEvent(aggregateID int64, aggregateType, eventType string, payload interface{}) (OutboxEvent, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return OutboxEvent{}, err
	}

	return OutboxEvent{
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		EventType:     eventType,
		EventPayload:  data,
		CreatedAt:     time.Now(),
		Status:        "pending",
		RetryCount:    0,
	}, nil
}
