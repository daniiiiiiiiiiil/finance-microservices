package ports

import (
	"context"

	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/repository/postgres/pool"
)

type UsersRepositoryInterface interface {
	GetUser(ctx context.Context, id int) (domain.User, error)
	DeleteUserTx(ctx context.Context, tx pool.Tx, id int) error
	PatchUser(ctx context.Context, id int, patch domain.User) (domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
	CreateUser(ctx context.Context, user domain.User) (int, error)
	ListUsers(ctx context.Context, limit, offset int) ([]domain.User, int, error)
	UpdateRole(ctx context.Context, id int, isAdmin bool) (domain.User, error)
	GetTotalUsers(ctx context.Context) (int, error)
	AdminExists(ctx context.Context) (bool, error)
	UpdateStatusTx(ctx context.Context, tx pool.Tx, id int, status string) error
}
