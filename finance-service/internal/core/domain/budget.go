package domain

import (
	"fmt"
	"time"
)

type BudgetPeriod string

const (
	BudgetPeriodDay   BudgetPeriod = "day"
	BudgetPeriodWeek  BudgetPeriod = "week"
	BudgetPeriodMonth BudgetPeriod = "month"
	BudgetPeriodYear  BudgetPeriod = "year"
)

type Budget struct {
	ID          int
	Version     int
	UserID      int
	Category    string
	AmountLimit float64
	Period      BudgetPeriod
	PeriodStart time.Time
	PeriodEnd   time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type BudgetStatus struct {
	Budget     Budget  `json:"budget"`
	Spent      float64 `json:"spent"`
	Remaining  float64 `json:"remaining"`
	Percentage float64 `json:"percentage"`
	Exceeded   bool    `json:"exceeded"`
}

func (b *Budget) Validate() error {
	if b.UserID <= 0 {
		return fmt.Errorf("user_id must be positive")
	}
	if b.Category == "" {
		return fmt.Errorf("category is required")
	}
	if b.AmountLimit <= 0 {
		return fmt.Errorf("amount_limit must be greater than zero")
	}
	switch b.Period {
	case BudgetPeriodDay, BudgetPeriodWeek, BudgetPeriodMonth, BudgetPeriodYear:
	default:
		return fmt.Errorf("invalid period %q", b.Period)
	}
	if !b.PeriodEnd.After(b.PeriodStart) {
		return fmt.Errorf("period_end must be after period_start")
	}
	return nil
}

func PeriodRange(now time.Time, p BudgetPeriod) (time.Time, time.Time, error) {
	now = now.UTC()
	switch p {
	case BudgetPeriodDay:
		from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		return from, from.AddDate(0, 0, 1), nil
	case BudgetPeriodWeek:
		daysSinceMonday := (int(now.Weekday()) + 6) % 7
		monday := now.AddDate(0, 0, -daysSinceMonday)
		from := time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, time.UTC)
		return from, from.AddDate(0, 0, 7), nil
	case BudgetPeriodMonth:
		from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		return from, from.AddDate(0, 1, 0), nil
	case BudgetPeriodYear:
		from := time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
		return from, from.AddDate(1, 0, 0), nil
	default:
		return time.Time{}, time.Time{}, fmt.Errorf("invalid period %q", p)
	}
}
