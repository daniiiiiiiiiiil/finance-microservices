package domain

import (
	errors_core "backend/internal/core/errors"
	"fmt"
	"time"
)

type Dashboard struct {
	TotalBalance    float64 `json:"total_balance"`
	MonthlyIncome   float64 `json:"monthly_income"`
	MonthlyExpenses float64 `json:"monthly_expenses"`
	SavingsRate     float64 `json:"savings_rate"`

	DailyStats []DailyStat `json:"daily_stats"`

	RecentTxs []RecentTransaction `json:"recent_transactions"`

	TopCategories []CategoryStat `json:"top_categories"`
}

type DailyStat struct {
	Date    string  `json:"date"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
}

type RecentTransaction struct {
	ID        int       `json:"id"`
	Type      string    `json:"type"`
	Amount    float64   `json:"amount"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"created_at"`
}

type CategoryStat struct {
	Category   string  `json:"category"`
	Total      float64 `json:"total"`
	Percentage float64 `json:"percentage"`
}

type Finance struct {
	ID              int
	Version         int
	TypeTransaction string
	Amount          float64
	Category        string
	CreatedAt       time.Time
	UserID          int
}

func NewFinance(id int, version int, typeTransaction string, amount float64, createdAt time.Time, userID int) *Finance {
	return &Finance{
		ID:              id,
		Version:         version,
		TypeTransaction: typeTransaction,
		Amount:          amount,
		CreatedAt:       createdAt,
		UserID:          userID,
	}
}

func (t Finance) Validate() error {
	if t.TypeTransaction != "income" && t.TypeTransaction != "expense" {
		return fmt.Errorf("invalid type: must be 'income' or 'expense'")
	}
	if t.Amount <= 0 {
		return fmt.Errorf("invalid amount %f : %w", t.Amount, errors_core.ErrInvalidArgument)
	}
	if t.Category == "" || len(t.Category) < 1 {
		return fmt.Errorf("invalid category len: %d : %w", len(t.Category), errors_core.ErrInvalidArgument)
	}
	if t.UserID < 0 {
		return fmt.Errorf("invalid user id %d : %w", t.UserID, errors_core.ErrInvalidArgument)
	}
	return nil
}
