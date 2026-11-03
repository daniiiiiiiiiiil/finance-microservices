package postgres

import (
	"encoding/json"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/core/domain"
)

type PlanModel struct {
	ID        int
	UserID    int
	RequestID string
	Text      string
	Address   string
	PlanJSON  []byte
	CreatedAt time.Time
}

func (m *PlanModel) ToDomain() (*domain.SavedPlan, error) {
	var plan domain.PlanResponse
	if err := json.Unmarshal(m.PlanJSON, &plan); err != nil {
		return nil, err
	}
	return &domain.SavedPlan{
		ID:        m.ID,
		UserID:    m.UserID,
		RequestID: m.RequestID,
		Text:      m.Text,
		Address:   m.Address,
		Plan:      plan,
		CreatedAt: m.CreatedAt,
	}, nil
}

func planFromDomain(p *domain.SavedPlan) (*PlanModel, error) {
	data, err := json.Marshal(p.Plan)
	if err != nil {
		return nil, err
	}
	return &PlanModel{
		ID:        p.ID,
		UserID:    p.UserID,
		RequestID: p.RequestID,
		Text:      p.Text,
		Address:   p.Address,
		PlanJSON:  data,
		CreatedAt: p.CreatedAt,
	}, nil
}
