package service_user

import (
	"backend/internal/core/cache"
	"backend/internal/core/domain"
	"backend/internal/core/kafka"
	"backend/internal/core/logger"
	"backend/internal/core/repository/postgres/pool"
	redisCache "backend/internal/features/users/repository/redis"
	"context"
)

type UsersService struct {
	userRepository UsersRepository
	pool           pool.Pool
	userCache      *redisCache.UserCache
	usersListCache *redisCache.UsersListCache
	producer       *kafka.Producer
	logger         *logger.Logger
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
	UpdateStatusTx(ctx context.Context, tx pool.Tx, id int, status string) error
}

func NewUsersService(userRepository UsersRepository, pool pool.Pool, redisClient *cache.RedisClient, producer *kafka.Producer, logger *logger.Logger) *UsersService {
	return &UsersService{
		userRepository: userRepository,
		pool:           pool,
		userCache:      redisCache.NewUserCache(redisClient),
		usersListCache: redisCache.NewUsersListCache(redisClient),
		producer:       producer,
		logger:         logger,
	}
}

type CreateProfileRequest struct {
	Email       string
	FullName    string
	PhoneNumber *string
	IsAdmin     bool
}
