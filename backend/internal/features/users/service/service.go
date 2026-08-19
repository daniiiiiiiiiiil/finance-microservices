package service_user

import (
	"backend/internal/core/cache"
	"backend/internal/core/domain"
	"backend/internal/core/repository/postgres/pool"
	redisCache "backend/internal/features/users/repository/redis"
	"context"
)

type UsersService struct {
	userRepository UsersRepository
	pool           pool.Pool
	userCache      *redisCache.UserCache
	usersListCache *redisCache.UsersListCache
}

//go:generate mockgen -destination=mocks/mock_users_service.go -package=mocks -source=service.go UsersService
type UsersRepository interface {
	GetUser(ctx context.Context, id int) (domain.User, error)
	DeleteUserTx(ctx context.Context, tx pool.Tx, id int) error
	PatchUser(ctx context.Context, id int, patch domain.User) (domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
	CreateUser(ctx context.Context, user domain.User) (int, error)
	ListUsers(ctx context.Context, limit, offset int) ([]domain.User, int, error)
	UpdateRoleUsers(ctx context.Context, id int, isAdmin bool) (domain.User, error)
	GetTotalUsers(ctx context.Context) (int, error)
	AdminExists(ctx context.Context) (bool, error)
}

func NewUsersService(userRepository UsersRepository, pool pool.Pool, redisClient *cache.RedisClient) *UsersService {
	return &UsersService{
		userRepository: userRepository,
		pool:           pool,
		userCache:      redisCache.NewUserCache(redisClient),
		usersListCache: redisCache.NewUsersListCache(redisClient)}
}

type CreateProfileRequest struct {
	Email       string
	FullName    string
	PhoneNumber *string
	IsAdmin     bool
}
