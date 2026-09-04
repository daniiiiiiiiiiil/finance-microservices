package service_user

import (
	"context"

	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/pkg/logger"
	"go.uber.org/zap"
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

func (s *UsersService) publishUserEvent(ctx context.Context, eventType string, userID int, email, fullName string, isAdmin bool, status string) {
	eventData := domain.UserEvent{
		UserID:   userID,
		Email:    email,
		FullName: fullName,
		IsAdmin:  isAdmin,
		Status:   status,
	}
	if err := s.eventPublisher.Publish(ctx, eventType, eventData); err != nil {
		s.logger.Warn("failed to publish user event", zap.Error(err))
	}
}
