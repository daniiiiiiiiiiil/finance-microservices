package service_admin

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	financeClient "github.com/daniiiiiiiiiiil/finance-microservices/admin-service/internal/core/clients/finance"
	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/internal/core/clients/users"
	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	userpb "github.com/daniiiiiiiiiiil/finance-microservices/admin-service/proto/users/gen"
)

type MockRedis struct {
	mock.Mock
}

func (m *MockRedis) Get(ctx context.Context, key string, dest interface{}) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

func (m *MockRedis) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockRedis) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockRedis) Incr(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRedis) Expire(ctx context.Context, key string, ttl time.Duration) error {
	args := m.Called(ctx, key, ttl)
	return args.Error(0)
}

func (m *MockRedis) Exists(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

type MockUsersClient struct {
	mock.Mock
}

func (m *MockUsersClient) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.UserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userpb.UserResponse), args.Error(1)
}

func (m *MockUsersClient) ListUsers(ctx context.Context, req *userpb.ListUsersRequest) (*users.ListUsersResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*users.ListUsersResponse), args.Error(1)
}

func (m *MockUsersClient) UpdateRole(ctx context.Context, req *userpb.UpdateRoleRequest) (*users.UserProfile, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*users.UserProfile), args.Error(1)
}

func (m *MockUsersClient) MarkDeleting(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUsersClient) RestoreUser(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUsersClient) FinalizeDelete(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUsersClient) GetMetrics(ctx context.Context) (*users.UserMetrics, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*users.UserMetrics), args.Error(1)
}

func (m *MockUsersClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

type MockFinanceClient struct {
	mock.Mock
}

func (m *MockFinanceClient) DeleteUserTransactions(ctx context.Context, userID int) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockFinanceClient) GetMetrics(ctx context.Context) (*financeClient.FinanceMetrics, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*financeClient.FinanceMetrics), args.Error(1)
}

func (m *MockFinanceClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) Publish(ctx context.Context, eventType string, data interface{}) error {
	args := m.Called(ctx, eventType, data)
	return args.Error(0)
}

type testSuite struct {
	service       *AdminService
	mockRedis     *MockRedis
	mockUsers     *MockUsersClient
	mockFinance   *MockFinanceClient
	mockPublisher *MockPublisher
}

func setup() *testSuite {
	mockRedis := new(MockRedis)
	mockUsers := new(MockUsersClient)
	mockFinance := new(MockFinanceClient)
	mockPublisher := new(MockPublisher)

	loggerInstance := &logger.Logger{Logger: zap.NewNop()}

	service := &AdminService{
		redis:          mockRedis,
		eventPublisher: mockPublisher,
		userClient:     mockUsers,
		financeClient:  mockFinance,
		logger:         loggerInstance,
	}

	return &testSuite{
		service:       service,
		mockRedis:     mockRedis,
		mockUsers:     mockUsers,
		mockFinance:   mockFinance,
		mockPublisher: mockPublisher,
	}
}

func TestGetUser_Success(t *testing.T) {
	s := setup()
	ctx := context.Background()

	user := domain.User{
		ID:       1,
		FullName: "John Doe",
		Email:    "john@example.com",
		IsAdmin:  false,
	}

	s.mockRedis.On("Get", ctx, "user:1", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*domain.User)
		*dest = user
	})

	result, err := s.service.GetUser(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, user, result)
	s.mockRedis.AssertExpectations(t)
}

func TestGetUser_NoCache(t *testing.T) {
	s := setup()
	ctx := context.Background()

	s.mockRedis.On("Get", ctx, "user:1", mock.Anything).Return(errors.New("not found"))

	s.mockUsers.On("GetUser", ctx, &userpb.GetUserRequest{Id: 1}).Return(&userpb.UserResponse{
		Id:       1,
		FullName: "John Doe",
		Email:    "john@example.com",
		IsAdmin:  false,
	}, nil)

	s.mockRedis.On("Set", ctx, "user:1", mock.Anything, 10*time.Minute).Return(nil)

	result, err := s.service.GetUser(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, 1, result.ID)
	assert.Equal(t, "John Doe", result.FullName)
	s.mockUsers.AssertExpectations(t)
}

func TestGetUsers_Success(t *testing.T) {
	s := setup()
	ctx := context.Background()

	s.mockUsers.On("ListUsers", ctx, &userpb.ListUsersRequest{Limit: 10, Offset: 0}).Return(&users.ListUsersResponse{
		Users: []users.UserProfile{
			{
				ID:       1,
				FullName: "John Doe",
				Email:    "john@example.com",
				IsAdmin:  false,
				IsActive: true,
			},
			{
				ID:       2,
				FullName: "Jane Smith",
				Email:    "jane@example.com",
				IsAdmin:  true,
				IsActive: true,
			},
		},
		Limit:  10,
		Offset: 0,
	}, nil)

	users, err := s.service.GetUsers(ctx, 10, 0)

	assert.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, 1, users[0].ID)
	assert.Equal(t, "Jane Smith", users[1].FullName)
	s.mockUsers.AssertExpectations(t)
}

func TestDeleteUser_Success(t *testing.T) {
	s := setup()
	ctx := context.Background()

	id := 2
	adminID := 1

	s.mockUsers.On("MarkDeleting", ctx, id).Return(nil)
	s.mockFinance.On("DeleteUserTransactions", ctx, id).Return(nil)
	s.mockUsers.On("FinalizeDelete", ctx, id).Return(nil)

	err := s.service.DeleteUser(ctx, id, adminID)

	assert.NoError(t, err)
	s.mockUsers.AssertExpectations(t)
	s.mockFinance.AssertExpectations(t)
}

func TestDeleteUser_SelfDeletion(t *testing.T) {
	s := setup()
	ctx := context.Background()

	id := 1
	adminID := 1

	err := s.service.DeleteUser(ctx, id, adminID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete themselves")
}

func TestDeleteUser_TransactionFailure(t *testing.T) {
	s := setup()
	ctx := context.Background()

	id := 2
	adminID := 1

	s.mockUsers.On("MarkDeleting", ctx, id).Return(nil)
	s.mockFinance.On("DeleteUserTransactions", ctx, id).Return(errors.New("transaction error"))
	s.mockUsers.On("RestoreUser", ctx, id).Return(nil)

	err := s.service.DeleteUser(ctx, id, adminID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "marking deleting user failed")
	s.mockUsers.AssertExpectations(t)
	s.mockFinance.AssertExpectations(t)
}

func TestUpdateUserRole_Success(t *testing.T) {
	s := setup()
	ctx := context.Background()

	id := 1
	isAdmin := true

	s.mockUsers.On("UpdateRole", ctx, &userpb.UpdateRoleRequest{Id: 1, IsAdmin: true}).Return(&users.UserProfile{
		ID:       1,
		Email:    "john@example.com",
		IsAdmin:  true,
		IsActive: true,
	}, nil)

	user, err := s.service.UpdateUserRole(ctx, id, isAdmin)

	assert.NoError(t, err)
	assert.Equal(t, 1, user.ID)
	assert.True(t, user.IsAdmin)
	s.mockUsers.AssertExpectations(t)
}

func TestUpdateUserRole_InvalidID(t *testing.T) {
	s := setup()
	ctx := context.Background()

	_, err := s.service.UpdateUserRole(ctx, 0, true)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Id must be greater than 0")
}

func (s *AdminService) invalidateCache(ctx context.Context, userID int) {
	if err := s.redis.Delete(ctx, fmt.Sprintf("user:%d", userID)); err != nil {
		s.logger.Warn("failed to invalidate user cache", zap.Int("user_id", userID), zap.Error(err))
	}
	if err := s.redis.Delete(ctx, "admin:metrics"); err != nil {
		s.logger.Warn("failed to invalidate metrics cache", zap.Error(err))
	}
	if err := s.redis.Delete(ctx, "admin:dashboard"); err != nil {
		s.logger.Warn("failed to invalidate dashboard cache", zap.Error(err))
	}
}

func TestInvalidateCache(t *testing.T) {
	s := setup()
	ctx := context.Background()

	s.mockRedis.On("Delete", ctx, "user:1").Return(nil)
	s.mockRedis.On("Delete", ctx, "admin:metrics").Return(nil)
	s.mockRedis.On("Delete", ctx, "admin:dashboard").Return(nil)

	s.service.invalidateCache(ctx, 1)

	s.mockRedis.AssertExpectations(t)
}

func TestDeleteUser_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		id          int
		adminID     int
		markErr     error
		delTxErr    error
		restoreErr  error
		finalizeErr error
		expectError bool
	}{
		{
			name:        "successful delete",
			id:          2,
			adminID:     1,
			expectError: false,
		},
		{
			name:        "cannot delete self",
			id:          1,
			adminID:     1,
			expectError: true,
		},
		{
			name:        "transaction delete failure",
			id:          2,
			adminID:     1,
			delTxErr:    errors.New("tx error"),
			expectError: true,
		},
		{
			name:        "finalize failure",
			id:          2,
			adminID:     1,
			finalizeErr: errors.New("finalize error"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := setup()
			ctx := context.Background()

			if tt.id != tt.adminID {
				s.mockUsers.On("MarkDeleting", ctx, tt.id).Return(tt.markErr)

				if tt.delTxErr == nil {
					s.mockFinance.On("DeleteUserTransactions", ctx, tt.id).Return(nil)

					if tt.finalizeErr == nil {
						s.mockUsers.On("FinalizeDelete", ctx, tt.id).Return(nil)
					} else {
						s.mockUsers.On("FinalizeDelete", ctx, tt.id).Return(tt.finalizeErr)
					}
				} else {
					s.mockFinance.On("DeleteUserTransactions", ctx, tt.id).Return(tt.delTxErr)
					s.mockUsers.On("RestoreUser", ctx, tt.id).Return(tt.restoreErr)
				}
			}

			err := s.service.DeleteUser(ctx, tt.id, tt.adminID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
