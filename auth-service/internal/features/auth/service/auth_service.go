package service_auth

import (
	"errors"

	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/core/cache"
	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/features/auth/repository/redis"
)

var _ ports.AuthServiceInterface = (*AuthService)(nil)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrRegistrationFailed = errors.New("registration failed")
)

type AuthService struct {
	credRepo    ports.CredRepoInterface
	jwtManager  ports.JWTManagerInterface
	redisCache  *cache.RedisClient
	rateLimit   ports.RateLimitInterface
	blacklist   ports.BlacklistInterface
	usersClient ports.UsersClientInterface
}

func NewAuthService(
	credRepo ports.CredRepoInterface,
	jwtManager ports.JWTManagerInterface,
	redisClient *cache.RedisClient,
	usersClient ports.UsersClientInterface,
) *AuthService {
	rateLimit := redis.NewRateLimitCache(redisClient)
	blacklist := redis.NewBlacklistCache(redisClient)

	return &AuthService{
		credRepo:    credRepo,
		jwtManager:  jwtManager,
		redisCache:  redisClient,
		rateLimit:   rateLimit,
		blacklist:   blacklist,
		usersClient: usersClient,
	}
}
