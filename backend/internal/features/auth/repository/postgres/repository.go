package postgres_auth

import (
	"backend/internal/core/repository/postgres/pool"
)

type AuthRepository struct {
	pool pool.Pool
}

func NewAuthRepository(pool pool.Pool) *AuthRepository {
	return &AuthRepository{pool: pool}
}
