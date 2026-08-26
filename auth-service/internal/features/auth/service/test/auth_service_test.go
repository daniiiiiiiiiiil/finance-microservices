package test

import (
	"backend/internal/core/auth/jwt"
	"backend/internal/core/cache"
	service_auth "backend/internal/features/auth/service"
	"backend/internal/features/auth/service/mocks"
	"backend/internal/features/auth/transport/http/dto"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func generateTestHash(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash)
}

func TestAuthService_Register_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockRedis := &cache.RedisClient{}
	jwtManager := jwt.NewJWTManager("test-secret", 24*time.Hour)

	service := service_auth.NewAuthService(mockRepo, jwtManager, mockRedis)

	ctx := context.Background()
	req := dto.RegisterRequest{
		FullName:    "Test User",
		Email:       "test@example.com",
		Password:    "password123",
		PhoneNumber: "+79991234567",
	}

	mockRepo.EXPECT().
		GetUserByEmail(ctx, req.Email).
		Return(service_auth.User{}, errors.New("not found")).
		Times(1)

	mockRepo.EXPECT().
		CreateUserWithAuth(ctx, req.FullName, req.Email, gomock.Any(), gomock.Any(), false).
		Return(1, nil).
		Times(1)

	token, user, err := service.Register(ctx, req)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, "Test User", user.FullName)
	assert.Equal(t, "test@example.com", user.Email)
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockRedis := &cache.RedisClient{}
	jwtManager := jwt.NewJWTManager("test-secret", 24*time.Hour)

	service := service_auth.NewAuthService(mockRepo, jwtManager, mockRedis)

	ctx := context.Background()
	req := dto.RegisterRequest{
		FullName:    "Test User",
		Email:       "test@example.com",
		Password:    "password123",
		PhoneNumber: "+79991234567",
	}

	mockRepo.EXPECT().
		GetUserByEmail(ctx, req.Email).
		Return(service_auth.User{ID: 1}, nil).
		Times(1)

	token, user, err := service.Register(ctx, req)

	assert.Error(t, err)
	assert.Equal(t, service_auth.ErrUserAlreadyExists, err)
	assert.Empty(t, token)
	assert.Nil(t, user)
}

func TestAuthService_Login_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockRedis := &cache.RedisClient{}
	jwtManager := jwt.NewJWTManager("test-secret", 24*time.Hour)

	service := service_auth.NewAuthService(mockRepo, jwtManager, mockRedis)

	ctx := context.Background()
	req := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	hashedPassword := generateTestHash("password123")

	mockRepo.EXPECT().
		GetUserByEmail(ctx, req.Email).
		Return(service_auth.User{ID: 1, PasswordHash: hashedPassword}, nil).
		Times(1)

	token, user, err := service.Login(ctx, req)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, 1, user.ID)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)
	mockRedis := &cache.RedisClient{}
	jwtManager := jwt.NewJWTManager("test-secret", 24*time.Hour)

	service := service_auth.NewAuthService(mockRepo, jwtManager, mockRedis)

	ctx := context.Background()
	req := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	hashedPassword := generateTestHash("password123")

	mockRepo.EXPECT().
		GetUserByEmail(ctx, req.Email).
		Return(service_auth.User{ID: 1, PasswordHash: hashedPassword}, nil).
		Times(1)

	token, user, err := service.Login(ctx, req)

	assert.Error(t, err)
	assert.Equal(t, service_auth.ErrInvalidCredentials, err)
	assert.Empty(t, token)
	assert.Nil(t, user)
}
