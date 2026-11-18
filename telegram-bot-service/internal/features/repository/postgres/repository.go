package postgres

import (
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/repository/postgres/pool"
)

type Repository struct {
	pool pool.Pool
}

func NewRepository(pool pool.Pool) *Repository {
	return &Repository{pool: pool}
}
