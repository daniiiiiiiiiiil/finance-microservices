package ports

import (
	"context"
	usersclient "github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/core/clients/users"
	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/core/domain"
	"time"
)

type RegisterRequest struct {
	FullName    string
	Email       string
	Password    string
	PhoneNumber string
	IsAdmin     bool
}

type LoginRequest struct {
	Email    string
	Password string
}

type UserResponse struct {
	ID          int
	FullName    string
	Email       string
	PhoneNumber *string
	IsAdmin     bool
}

type AuthServiceInterface interface {
	Register(ctx context.Context, req RegisterRequest) (string, *UserResponse, error)
	Login(ctx context.Context, req LoginRequest) (string, *UserResponse, error)
	GenerateToken(userID int, email string, isAdmin bool) (string, error)
	AdminExists(ctx context.Context) (bool, error)
	RateLimitCheck(ctx context.Context, key string, limit int64, ttl time.Duration) (bool, error)
	AddToBlacklist(ctx context.Context, token string, ttl time.Duration) error
}

type CredRepoInterface interface {
	GetByEmail(ctx context.Context, email string) (*domain.Credentials, error)
	Create(ctx context.Context, email, passwordHash string) (int, error)
	AdminUpdateStatus(ctx context.Context, id int, status string) error
}

type UsersClientInterface interface {
	CreateProfile(ctx context.Context, req *usersclient.CreateProfileRequest) (*usersclient.UserProfile, error)
	GetUserByEmail(ctx context.Context, email string) (*usersclient.UserProfile, error)
	AdminExists(ctx context.Context) (bool, error)
	Close() error
}
