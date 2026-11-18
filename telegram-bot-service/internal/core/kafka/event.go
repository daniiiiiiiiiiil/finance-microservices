package kafka

import (
	"encoding/json"
	"time"
)

type Event struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

type BudgetExceededEvent struct {
	BudgetID    int       `json:"budget_id"`
	UserID      int       `json:"user_id"`
	Category    string    `json:"category"`
	Period      string    `json:"period"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	AmountLimit float64   `json:"amount_limit"`
	Spent       float64   `json:"spent"`
	ExceededBy  float64   `json:"exceeded_by"`
}

type PlanCreatedEvent struct {
	PlanID    int       `json:"plan_id"`
	UserID    int       `json:"user_id"`
	RequestID string    `json:"request_id"`
	TotalCost float64   `json:"total_cost"`
	Mode      string    `json:"mode"`
	CreatedAt time.Time `json:"created_at"`
}

type TransactionEvent struct {
	TransactionID   int       `json:"transaction_id"`
	UserID          int       `json:"user_id"`
	TypeTransaction string    `json:"type_transaction"`
	Amount          float64   `json:"amount"`
	Category        string    `json:"category"`
	CreatedAt       time.Time `json:"created_at"`
}

type UserDeletedEvent struct {
	UserID    int       `json:"user_id"`
	Email     string    `json:"email"`
	DeletedAt time.Time `json:"deleted_at"`
}
