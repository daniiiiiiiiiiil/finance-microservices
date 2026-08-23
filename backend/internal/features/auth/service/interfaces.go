package service_auth

import (
	"time"

	"golang.org/x/net/context"
)

type AuthServiceInterface interface {
	Register(ctx context.Context, req RegisterRequest) (string, *UserResponse, error)
	Login(ctx context.Context, req LoginRequest) (string, *UserResponse, error)
	GenerateToken(userID int, email string, isAdmin bool) (string, error)
	AdminExists(ctx context.Context) (bool, error)
	RateLimitCheck(ctx context.Context, key string, limit int64, ttl time.Duration) (bool, error)
	AddToBlacklist(ctx context.Context, token string, ttl time.Duration) error
}
