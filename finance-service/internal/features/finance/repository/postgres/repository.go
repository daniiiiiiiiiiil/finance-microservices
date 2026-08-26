package postgres

import "github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/repository/postgres/pool"

type FinanceRepository struct {
	pool pool.Pool
}

func NewFinanceRepository(p pool.Pool) *FinanceRepository {
	return &FinanceRepository{pool: p}
}
