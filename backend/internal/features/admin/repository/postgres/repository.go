package postgres

import "backend/internal/core/repository/postgres/pool"

type AdminRepository struct {
	pool pool.Pool
}

func NewAdminRepository(p pool.Pool) *AdminRepository {
	return &AdminRepository{pool: p}
}
