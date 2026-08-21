package kafka

import (
	"encoding/json"
	"time"
)

const (
	EventTypeTransactionCreated      = "transaction.created"
	EventTypeTransactionUpdated      = "transaction.updated"
	EventTypeTransactionDeleted      = "transaction.deleted"
	EventTypeUserCreated             = "user.created"
	EventTypeUserDeleted             = "user.deleted"
	EventTypeAdminMetrics            = "admin.metrics"
	EventTypeUserTransactionsDeleted = "user.transactions.deleted"
	EventTypeCurrencyCreate          = "create.currency"
	EventTypeCurrencyDelete          = "delete.currency"
)

type Event struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

type TransactionEvent struct {
	TransactionID   int       `json:"transaction_id"`
	UserID          int       `json:"user_id"`
	TypeTransaction string    `json:"type_transaction"`
	Amount          float64   `json:"amount"`
	Category        string    `json:"category"`
	CreatedAt       time.Time `json:"created_at"`
}

type UserEvent struct {
	UserID   int    `json:"user_id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	IsAdmin  bool   `json:"is_admin"`
	Status   string `json:"status"`
}

type ConvertedEvent struct {
	TransactionID    int       `json:"transaction_id"`
	AmountUSD        float64   `json:"amount_usd"`
	OriginalAmount   float64   `json:"original_amount"`
	OriginalCurrency string    `json:"original_currency"`
	ConvertedAt      time.Time `json:"converted_at"`
}

type RatesUpdatedEvent struct {
	Base      string             `json:"base"`
	Rates     map[string]float64 `json:"rates"`
	UpdatedAt time.Time          `json:"updated_at"`
}

type TransactionCurrencyEvent struct {
	TransactionID int       `json:"transaction_id"`
	UserID        int       `json:"user_id"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Category      string    `json:"category"`
	CreatedAt     time.Time `json:"created_at"`
}

type MetricsEvent struct {
	TotalUsers        int       `json:"total_users"`
	TotalTransactions int       `json:"total_transactions"`
	TotalBalance      float64   `json:"total_balance"`
	Timestamp         time.Time `json:"timestamp"`
}

type UserTransactionsDeletedEvent struct {
	UserID       int       `json:"user_id"`
	DeletedCount int       `json:"deleted_count"`
	Timestamp    time.Time `json:"timestamp"`
}
