package service_auth

import (
	"backend/internal/core/auth/jwt"
	"backend/internal/core/cache"
	"backend/internal/features/auth/repository/redis"
	"context"
	"errors"
	"time"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type AuthService struct {
	userRepo   UserRepository
	jwtManager *jwt.JWTManager
	redisCache *cache.RedisClient
	rateLimit  *redis.RateLimitCache
	blacklist  *redis.BlacklistCache
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

func NewAuthService(userRepo UserRepository, jwtManager *jwt.JWTManager, redisCache *cache.RedisClient) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
		redisCache: redisCache,
		rateLimit:  redis.NewRateLimitCache(redisCache),
		blacklist:  redis.NewBlacklistCache(redisCache),
	}
}

func (s *AuthService) GenerateToken(ctx context.Context, userID int) (string, error) {
	return s.jwtManager.Generate(userID)
}

func (s *AuthService) AdminExists(ctx context.Context) (bool, error) {
	return s.userRepo.AdminExists(ctx)
}

func (s *AuthService) RateLimitCheck(ctx context.Context, key string, limit int64, ttl time.Duration) (bool, error) {
	return s.rateLimit.Check(ctx, key, limit, ttl)
}

func (s *AuthService) AddToBlacklist(ctx context.Context, token string, ttl time.Duration) error {
	return s.blacklist.Add(ctx, token, ttl)
}
