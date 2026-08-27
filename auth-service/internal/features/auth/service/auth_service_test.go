package service_auth

import (
	"context"
	"errors"
	"testing"
	"time"

	usersclient "github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/core/clients/users"
	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/core/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type MockCredRepo struct {
	mock.Mock
}

func (m *MockCredRepo) GetByEmail(ctx context.Context, email string) (*domain.Credentials, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Credentials), args.Error(1)
}

func (m *MockCredRepo) Create(ctx context.Context, email, passwordHash string) (int, error) {
	args := m.Called(ctx, email, passwordHash)
	return args.Int(0), args.Error(1)
}

func (m *MockCredRepo) AdminUpdateStatus(ctx context.Context, id int, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

type MockUsersClient struct {
	mock.Mock
}

func (m *MockUsersClient) CreateProfile(ctx context.Context, req *usersclient.CreateProfileRequest) (*usersclient.UserProfile, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usersclient.UserProfile), args.Error(1)
}

func (m *MockUsersClient) GetUserByEmail(ctx context.Context, email string) (*usersclient.UserProfile, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usersclient.UserProfile), args.Error(1)
}

func (m *MockUsersClient) AdminExists(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}

func (m *MockUsersClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

type MockJWTManager struct {
	mock.Mock
}

func (m *MockJWTManager) Generate(userID int, email string, isAdmin bool) (string, error) {
	args := m.Called(userID, email, isAdmin)
	return args.String(0), args.Error(1)
}

type MockRateLimit struct {
	mock.Mock
}

func (m *MockRateLimit) Check(ctx context.Context, key string, limit int64, ttl time.Duration) (bool, error) {
	args := m.Called(ctx, key, limit, ttl)
	return args.Bool(0), args.Error(1)
}

func (m *MockRateLimit) Reset(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

type MockBlacklist struct {
	mock.Mock
}

func (m *MockBlacklist) Add(ctx context.Context, token string, ttl time.Duration) error {
	args := m.Called(ctx, token, ttl)
	return args.Error(0)
}

func (m *MockBlacklist) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	args := m.Called(ctx, token)
	return args.Bool(0), args.Error(1)
}

func (m *MockBlacklist) Remove(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

type testSuite struct {
	service       *AuthService
	mockCred      *MockCredRepo
	mockUsers     *MockUsersClient
	mockJWT       *MockJWTManager
	mockRateLimit *MockRateLimit
	mockBlacklist *MockBlacklist
}

func setup() *testSuite {
	mockCred := new(MockCredRepo)
	mockUsers := new(MockUsersClient)
	mockJWT := new(MockJWTManager)
	mockRateLimit := new(MockRateLimit)
	mockBlacklist := new(MockBlacklist)

	service := &AuthService{
		credRepo:    mockCred,
		jwtManager:  mockJWT,
		rateLimit:   mockRateLimit,
		blacklist:   mockBlacklist,
		usersClient: mockUsers,
	}

	return &testSuite{
		service:       service,
		mockCred:      mockCred,
		mockUsers:     mockUsers,
		mockJWT:       mockJWT,
		mockRateLimit: mockRateLimit,
		mockBlacklist: mockBlacklist,
	}
}

func TestRegister_Success(t *testing.T) {
	s := setup()
	ctx := context.Background()

	req := ports.RegisterRequest{
		FullName: "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}

	s.mockCred.On("GetByEmail", ctx, req.Email).Return(nil, nil)
	s.mockCred.On("Create", ctx, req.Email, mock.AnythingOfType("string")).Return(1, nil)
	s.mockCred.On("AdminUpdateStatus", ctx, 1, "active").Return(nil)

	s.mockUsers.On("CreateProfile", ctx, mock.Anything).Return(&usersclient.UserProfile{
		ID:      1,
		Email:   req.Email,
		IsAdmin: false,
	}, nil)

	s.mockJWT.On("Generate", 1, req.Email, false).Return("token123", nil)

	token, user, err := s.service.Register(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, "token123", token)
	assert.NotNil(t, user)
	assert.Equal(t, 1, user.ID)

	s.mockCred.AssertExpectations(t)
	s.mockUsers.AssertExpectations(t)
	s.mockJWT.AssertExpectations(t)
}

func TestRegister_UserAlreadyExists(t *testing.T) {
	s := setup()
	ctx := context.Background()

	req := ports.RegisterRequest{
		Email:    "john@example.com",
		Password: "password123",
	}

	s.mockCred.On("GetByEmail", ctx, req.Email).Return(&domain.Credentials{
		ID:    1,
		Email: req.Email,
	}, nil)

	token, _, err := s.service.Register(ctx, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
	assert.Empty(t, token)
}

func TestLogin_Success(t *testing.T) {
	s := setup()
	ctx := context.Background()

	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	req := ports.LoginRequest{
		Email:    "john@example.com",
		Password: password,
	}

	s.mockCred.On("GetByEmail", ctx, req.Email).Return(&domain.Credentials{
		ID:           1,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Status:       "active",
	}, nil)

	s.mockUsers.On("GetUserByEmail", ctx, req.Email).Return(&usersclient.UserProfile{
		ID:      1,
		Email:   req.Email,
		IsAdmin: false,
	}, nil)

	s.mockJWT.On("Generate", 1, req.Email, false).Return("token123", nil)

	token, user, err := s.service.Login(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, "token123", token)
	assert.NotNil(t, user)
	assert.Equal(t, 1, user.ID)
}

func TestLogin_WrongPassword(t *testing.T) {
	s := setup()
	ctx := context.Background()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)

	req := ports.LoginRequest{
		Email:    "john@example.com",
		Password: "wrongpassword",
	}

	s.mockCred.On("GetByEmail", ctx, req.Email).Return(&domain.Credentials{
		ID:           1,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Status:       "active",
	}, nil)

	_, _, err := s.service.Login(ctx, req)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
}

func TestLogin_UserNotFound(t *testing.T) {
	s := setup()
	ctx := context.Background()

	req := ports.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	s.mockCred.On("GetByEmail", ctx, req.Email).Return(nil, errors.New("not found"))

	_, _, err := s.service.Login(ctx, req)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCredentials, err)
}

func TestLogin_InactiveAccount(t *testing.T) {
	s := setup()
	ctx := context.Background()

	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	req := ports.LoginRequest{
		Email:    "john@example.com",
		Password: password,
	}

	s.mockCred.On("GetByEmail", ctx, req.Email).Return(&domain.Credentials{
		ID:           1,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Status:       "inactive",
	}, nil)

	_, _, err := s.service.Login(ctx, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "account not active")
}

func TestGenerateToken(t *testing.T) {
	s := setup()

	s.mockJWT.On("Generate", 1, "test@example.com", true).Return("token123", nil)

	token, err := s.service.GenerateToken(1, "test@example.com", true)

	assert.NoError(t, err)
	assert.Equal(t, "token123", token)
	s.mockJWT.AssertExpectations(t)
}

func TestAdminExists(t *testing.T) {
	s := setup()
	ctx := context.Background()

	s.mockUsers.On("AdminExists", ctx).Return(true, nil)

	exists, err := s.service.AdminExists(ctx)

	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestRateLimitCheck(t *testing.T) {
	s := setup()
	ctx := context.Background()

	key := "test-key"
	limit := int64(5)
	ttl := time.Minute

	s.mockRateLimit.On("Check", ctx, key, limit, ttl).Return(true, nil)

	allowed, err := s.service.RateLimitCheck(ctx, key, limit, ttl)

	assert.NoError(t, err)
	assert.True(t, allowed)
	s.mockRateLimit.AssertExpectations(t)
}

func TestAddToBlacklist(t *testing.T) {
	s := setup()
	ctx := context.Background()

	token := "test-token"
	ttl := time.Hour

	s.mockBlacklist.On("Add", ctx, token, ttl).Return(nil)

	err := s.service.AddToBlacklist(ctx, token, ttl)

	assert.NoError(t, err)
	s.mockBlacklist.AssertExpectations(t)
}

func TestLogin_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		password    string
		status      string
		expectError bool
	}{
		{
			name:        "successful login",
			email:       "user@example.com",
			password:    "password123",
			status:      "active",
			expectError: false,
		},
		{
			name:        "inactive account",
			email:       "user@example.com",
			password:    "password123",
			status:      "inactive",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := setup()
			ctx := context.Background()

			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(tt.password), bcrypt.DefaultCost)

			req := ports.LoginRequest{
				Email:    tt.email,
				Password: tt.password,
			}

			s.mockCred.On("GetByEmail", ctx, tt.email).Return(&domain.Credentials{
				ID:           1,
				Email:        tt.email,
				PasswordHash: string(hashedPassword),
				Status:       tt.status,
			}, nil)

			if tt.status == "active" {
				s.mockUsers.On("GetUserByEmail", ctx, tt.email).Return(&usersclient.UserProfile{
					ID:      1,
					Email:   tt.email,
					IsAdmin: false,
				}, nil)
				s.mockJWT.On("Generate", 1, tt.email, false).Return("token", nil)
			}

			_, _, err := s.service.Login(ctx, req)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
