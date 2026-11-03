package ports

import (
	"context"

	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/core/domain"
)

type PlanServiceInterface interface {
	PlanShopping(ctx context.Context, req PlanRequest) (*domain.PlanResponse, error)
	ListPlans(ctx context.Context, userID, limit, offset int) ([]domain.PlanSummary, int, error)
	GetPlan(ctx context.Context, userID, id int) (*domain.SavedPlan, error)
	DeletePlan(ctx context.Context, userID, id int) error
	DeletePlansByUser(ctx context.Context, userID int) error
}

type PlanRequest struct {
	RequestID    string
	UserID       int
	Text         string
	StartAddress string
}
