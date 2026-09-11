package postgres

import (
	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/repository/postgres/pool"
)

type SagaRepository struct {
	pool pool.Pool
}

func NewSagaRepository(pool pool.Pool) *SagaRepository {
	return &SagaRepository{pool: pool}
}
