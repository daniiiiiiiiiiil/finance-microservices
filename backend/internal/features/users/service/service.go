package service_user

import (
	"backend/internal/core/domain"
	"backend/internal/core/repository/postgres/pool"
	"context"
)

type UsersService struct {
	userRepository UsersRepository
	pool           pool.Pool
}

//go:generate mockgen -destination=mocks/mock_users_service.go -package=mocks -source=service.go UsersService
type UsersRepository interface {
	GetUser(ctx context.Context, id int) (domain.User, error)
	DeleteUserTx(ctx context.Context, tx pool.Tx, id int) error
	PatchUser(ctx context.Context, id int, patch domain.User) (domain.User, error)
}

func NewUsersService(userRepository UsersRepository, pool pool.Pool) *UsersService {
	return &UsersService{userRepository: userRepository, pool: pool}
}
