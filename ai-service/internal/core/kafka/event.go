package kafka

import (
	"encoding/json"
	"time"
)

const (
	EventTypePlanCreated = "ai.plan_created"
	EventTypePlanDeleted = "ai.plan_deleted"
	EventTypeUserDeleted = "user.deleted"
)

type Event struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

type PlanCreatedEvent struct {
	PlanID    int       `json:"plan_id"`
	UserID    int       `json:"user_id"`
	RequestID string    `json:"request_id"`
	TotalCost float64   `json:"total_cost"`
	Mode      string    `json:"mode"`
	CreatedAt time.Time `json:"created_at"`
}

type PlanDeletedEvent struct {
	PlanID int       `json:"plan_id"`
	UserID int       `json:"user_id"`
	At     time.Time `json:"at"`
}

type UserDeletedEvent struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
}
