package postgres_auth

import (
	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/core/repository/postgres/pool"
)

type AuthRepository struct {
	pool pool.Pool
}

func NewAuthRepository(pool pool.Pool) *AuthRepository {
	return &AuthRepository{pool: pool}
}
