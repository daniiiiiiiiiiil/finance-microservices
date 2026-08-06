package postgres

import (
	"backend/internal/core/domain"
	"time"
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

func financeModelsFromDomains(transactions []domain.Finance) []FinanceModel {
	models := make([]FinanceModel, len(transactions))
	for i, t := range transactions {
		models[i] = FinanceModel{
			ID:              t.ID,
			Version:         t.Version,
			TypeTransaction: t.TypeTransaction,
			Amount:          t.Amount,
			Category:        t.Category,
			CreatedAt:       t.CreatedAt,
			UserID:          t.UserID,
		}
	}
	return models
}
