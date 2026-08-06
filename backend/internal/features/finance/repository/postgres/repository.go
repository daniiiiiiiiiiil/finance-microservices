package postgres

import "backend/internal/core/repository/postgres/pool"

type FinanceRepository struct {
	pool pool.Pool
}

func NewFinanceRepository(p pool.Pool) *FinanceRepository {
	return &FinanceRepository{pool: p}
}
