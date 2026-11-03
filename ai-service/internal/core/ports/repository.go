package ports

import (
	"context"

	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/core/domain"
)

type PlanRepository interface {
	Save(ctx context.Context, plan *domain.SavedPlan) error
	GetByID(ctx context.Context, id int) (*domain.SavedPlan, error)
	GetByRequestID(ctx context.Context, requestID string) (*domain.SavedPlan, error)
	ListByUser(ctx context.Context, userID int, limit, offset int) ([]domain.PlanSummary, error)
	CountByUser(ctx context.Context, userID int) (int, error)
	Delete(ctx context.Context, id int) error
	DeleteByUser(ctx context.Context, userID int) error
}
