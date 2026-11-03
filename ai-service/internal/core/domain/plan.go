package domain

import (
	"errors"
	"time"
)

var (
	ErrPlanNotFound = errors.New("plan not found")
)

type SavedPlan struct {
	ID        int          `json:"id"`
	UserID    int          `json:"user_id"`
	RequestID string       `json:"request_id"`
	Text      string       `json:"text"`
	Address   string       `json:"address"`
	Plan      PlanResponse `json:"plan"`
	CreatedAt time.Time    `json:"created_at"`
}

type PlanSummary struct {
	ID          int       `json:"id"`
	RequestID   string    `json:"request_id"`
	Text        string    `json:"text"`
	Address     string    `json:"address"`
	ItemsCount  int       `json:"items_count"`
	StoresCount int       `json:"stores_count"`
	TotalCost   float64   `json:"total_cost"`
	Recommended string    `json:"recommended"`
	CreatedAt   time.Time `json:"created_at"`
}
