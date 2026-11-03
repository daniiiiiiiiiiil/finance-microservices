package postgres

import (
	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/core/repository/postgres/pool"
)

type PlanRepository struct {
	pool pool.Pool
}

func NewPlanRepository(p pool.Pool) *PlanRepository {
	return &PlanRepository{pool: p}
}
