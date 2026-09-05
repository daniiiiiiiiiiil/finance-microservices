package postgres

import (
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/domain"
)

type ShoppingModel struct {
	ID             uint       `json:"id"`
	Version        uint       `json:"version"`
	Title          string     `json:"title"`
	Description    *string    `json:"description"`
	AmountNow      float64    `json:"amount_now"`
	AmountFinish   float64    `json:"amount_finish"`
	ImageKey       *string    `json:"image_key"`
	Completed      bool       `json:"completed"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at"`
	CompletedAt    *time.Time `json:"completed_at"`
	CompletionDate *time.Time `json:"completion_date"`
}

func shoppingDomainFromModel(model ShoppingModel) domain.Shopping {
	return domain.Shopping{
		ID:             model.ID,
		Version:        model.Version,
		Title:          model.Title,
		Description:    model.Description,
		AmountNow:      model.AmountNow,
		AmountFinish:   model.AmountFinish,
		ImageKey:       model.ImageKey,
		Completed:      model.Completed,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
		CompletedAt:    model.CompletedAt,
		CompletionDate: model.CompletionDate,
	}
}
