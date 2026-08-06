package postgres

import "backend/internal/core/repository/postgres/pool"

type UserRepository struct {
	pool pool.Pool
}

func NewUserRepository(
	pool pool.Pool) *UserRepository {
	return &UserRepository{
		pool: pool,
	}
}
