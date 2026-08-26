package service_auth

import (
	"context"
	"errors"

	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/core/auth/jwt"
	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/core/cache"
	usersclient "github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/core/clients/users"
	postgres_auth "github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/features/auth/repository/postgres"
	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/features/auth/repository/redis"
)

var _ AuthServiceInterface = (*AuthService)(nil)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrRegistrationFailed = errors.New("registration failed")
)

type AuthService struct {
	credRepo    *postgres_auth.AuthRepository
	jwtManager  *jwt.JWTManager
	redisCache  *cache.RedisClient
	rateLimit   *redis.RateLimitCache
	blacklist   *redis.BlacklistCache
	usersClient *usersclient.UsersClient
}

//go:generate mockgen -destination=mocks/mock_auth_service.go -package=mocks -source=auth_service.go AuthService
type UserRepository interface {
	CreateUserWithAuth(ctx context.Context, fullName, email, passwordHash string, phoneNumber *string, isAdmin bool) (int, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	AdminExists(ctx context.Context) (bool, error)
}

type User struct {
	ID           int
	FullName     string
	Email        string
	PasswordHash string
	PhoneNumber  *string
	IsAdmin      bool
}

func NewAuthService(
	credRepo *postgres_auth.AuthRepository,
	jwtManager *jwt.JWTManager,
	redisCache *cache.RedisClient,
	usersClient *usersclient.UsersClient,
) *AuthService {
	return &AuthService{
		credRepo:    credRepo,
		jwtManager:  jwtManager,
		redisCache:  redisCache,
		rateLimit:   redis.NewRateLimitCache(redisCache),
		blacklist:   redis.NewBlacklistCache(redisCache),
		usersClient: usersClient,
	}
}
