package postgres

import (
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/domain"
)

type FinanceModel struct {
	ID              int
	Version         int
	TypeTransaction string
	Amount          float64
	Category        string
	CreatedAt       time.Time
	UserID          int
}

type BudgetModel struct {
	ID          int
	Version     int
	UserID      int
	Category    string
	AmountLimit float64
	Period      string
	PeriodStart time.Time
	PeriodEnd   time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func financeDomainFromModel(model FinanceModel) domain.Finance {
	return domain.Finance{
		ID:              model.ID,
		Version:         model.Version,
		TypeTransaction: model.TypeTransaction,
		Amount:          model.Amount,
		Category:        model.Category,
		CreatedAt:       model.CreatedAt,
		UserID:          model.UserID,
	}
}

func budgetDomainFromModel(m BudgetModel) domain.Budget {
	return domain.Budget{
		ID:          m.ID,
		Version:     m.Version,
		UserID:      m.UserID,
		Category:    m.Category,
		AmountLimit: m.AmountLimit,
		Period:      domain.BudgetPeriod(m.Period),
		PeriodStart: m.PeriodStart,
		PeriodEnd:   m.PeriodEnd,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
