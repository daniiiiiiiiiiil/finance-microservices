package postgres

import "github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/repository/postgres/pool"

type UserRepository struct {
	pool pool.Pool
}

func NewUserRepository(
	pool pool.Pool) *UserRepository {
	return &UserRepository{
		pool: pool,
	}
}
