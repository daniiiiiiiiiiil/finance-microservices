package http

import (
	"backend/internal/core/domain"
	"time"
)

type CreateTransactionRequest struct {
	TypeTransaction string  `json:"type_transaction" validate:"required,oneof=income expense"`
	Amount          float64 `json:"amount" validate:"required,gt=0"`
	Category        string  `json:"category" validate:"required,min=1,max=60"`
}

type UpdateTransactionRequest struct {
	TypeTransaction string  `json:"type_transaction" validate:"required,oneof=income expense"`
	Amount          float64 `json:"amount" validate:"required,gt=0"`
	Category        string  `json:"category" validate:"required,min=1,max=60"`
}

type TransactionResponse struct {
	ID              int       `json:"id"`
	Version         int       `json:"version"`
	TypeTransaction string    `json:"type_transaction"`
	Amount          float64   `json:"amount"`
	Category        string    `json:"category"`
	CreatedAt       time.Time `json:"created_at"`
	UserID          int       `json:"user_id"`
}

func transactionResponseFromDomain(t domain.Finance) TransactionResponse {
	return TransactionResponse{
		ID:              t.ID,
		Version:         t.Version,
		TypeTransaction: t.TypeTransaction,
		Amount:          t.Amount,
		Category:        t.Category,
		CreatedAt:       t.CreatedAt,
		UserID:          t.UserID,
	}
}

func transactionResponsesFromDomains(transactions []domain.Finance) []TransactionResponse {
	responses := make([]TransactionResponse, len(transactions))
	for i, t := range transactions {
		responses[i] = transactionResponseFromDomain(t)
	}
	return responses
}
