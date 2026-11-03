package service_user

import (
	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/pkg/logger"
)

type UsersService struct {
	userRepository ports.UsersRepositoryInterface
	pool           ports.PoolInterface
	userCache      ports.UserCacheInterface
	usersListCache ports.UsersListCacheInterface
	eventPublisher ports.EventPublisherInterface
	outboxRepo     ports.OutboxRepositoryInterface
	logger         *logger.Logger
	redis          ports.RedisInterface
}

func NewUsersService(
	userRepository ports.UsersRepositoryInterface,
	pool ports.PoolInterface,
	userCache ports.UserCacheInterface,
	usersListCache ports.UsersListCacheInterface,
	eventPublisher ports.EventPublisherInterface,
	outboxRepo ports.OutboxRepositoryInterface,
	logger *logger.Logger,
	redis ports.RedisInterface,
) *UsersService {
	return &UsersService{
		userRepository: userRepository,
		pool:           pool,
		userCache:      userCache,
		usersListCache: usersListCache,
		eventPublisher: eventPublisher,
		outboxRepo:     outboxRepo,
		logger:         logger,
		redis:          redis,
	}
}

type CreateProfileRequest struct {
	Email        string
	FullName     string
	PhoneNumber  *string
	IsAdmin      bool
	PasswordHash string
}
