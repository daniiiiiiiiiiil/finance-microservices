package domain

import (
	"time"
)

type TransactionEvent struct {
	TransactionID   int       `json:"transaction_id"`
	UserID          int       `json:"user_id"`
	TypeTransaction string    `json:"type_transaction"`
	Amount          float64   `json:"amount"`
	Category        string    `json:"category"`
	CreatedAt       time.Time `json:"created_at"`
}

type UserTransactionsDeletedEvent struct {
	UserID       int       `json:"user_id"`
	DeletedCount int       `json:"deleted_count"`
	Timestamp    time.Time `json:"timestamp"`
}
