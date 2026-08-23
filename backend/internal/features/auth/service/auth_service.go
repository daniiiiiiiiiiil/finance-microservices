package service_auth

import (
	"backend/internal/core/auth/jwt"
	"backend/internal/core/cache"
	usersclient "backend/internal/core/clients/users"
	postgres_auth "backend/internal/features/auth/repository/postgres"
	"backend/internal/features/auth/repository/redis"
	"context"
	"errors"
	"time"
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
func (s *AuthService) GenerateToken(userID int, email string, isAdmin bool) (string, error) {
	return s.jwtManager.Generate(userID, email, isAdmin)
}

func (s *AuthService) AdminExists(ctx context.Context) (bool, error) {
	exists, err := s.usersClient.AdminExists(ctx)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (s *AuthService) RateLimitCheck(ctx context.Context, key string, limit int64, ttl time.Duration) (bool, error) {
	return s.rateLimit.Check(ctx, key, limit, ttl)
}

func (s *AuthService) AddToBlacklist(ctx context.Context, token string, ttl time.Duration) error {
	return s.blacklist.Add(ctx, token, ttl)
}
