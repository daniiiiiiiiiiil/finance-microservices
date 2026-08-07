package test

import (
	"backend/internal/core/cache"
	"backend/internal/core/domain"
	"backend/internal/core/repository/postgres/pool"
	service_admin "backend/internal/features/admin/service"
	"backend/internal/features/admin/service/mocks"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockPool struct{}

func (m *MockPool) Query(ctx context.Context, sql string, args ...any) (pool.Rows, error) {
	return nil, nil
}

func (m *MockPool) QueryRow(ctx context.Context, sql string, args ...any) pool.Row {
	return nil
}

func (m *MockPool) Exec(ctx context.Context, sql string, args ...any) (pool.CommandTag, error) {
	return nil, nil
}

func (m *MockPool) Begin(ctx context.Context) (pool.Tx, error) {
	return &MockTx{}, nil
}

func (m *MockPool) Close() {}

func (m *MockPool) OpTimeout() time.Duration {
	return 5 * time.Second
}

type MockTx struct{}

func (m *MockTx) Exec(ctx context.Context, sql string, args ...any) (pool.CommandTag, error) {
	return nil, nil
}

func (m *MockTx) Query(ctx context.Context, sql string, args ...any) (pool.Rows, error) {
	return nil, nil
}

func (m *MockTx) QueryRow(ctx context.Context, sql string, args ...any) pool.Row {
	return nil
}

func (m *MockTx) Commit(ctx context.Context) error {
	return nil
}

func (m *MockTx) Rollback(ctx context.Context) error {
	return nil
}

type MockRedisClient struct {
	shouldReturnNil bool
}

func NewMockRedisClient(shouldReturnNil bool) *MockRedisClient {
	return &MockRedisClient{shouldReturnNil: shouldReturnNil}
}

func (m *MockRedisClient) Get(ctx context.Context, key string, dest interface{}) error {
	if m.shouldReturnNil {
		return redis.Nil
	}
	return nil
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return nil
}

func (m *MockRedisClient) Delete(ctx context.Context, key string) error {
	return nil
}

func (m *MockRedisClient) Incr(ctx context.Context, key string) (int64, error) {
	return 0, nil
}

func (m *MockRedisClient) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return nil
}

var _ cache.RedisInterface = (*MockRedisClient)(nil)

func TestAdminService_GetUsers_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAdminRepository(ctrl)
	mockPool := &MockPool{}
	mockRedis := NewMockRedisClient(false)

	service := service_admin.NewAdminService(mockRepo, mockPool, mockRedis)

	ctx := context.Background()
	expectedUsers := []domain.User{
		{ID: 1, FullName: "User 1"},
		{ID: 2, FullName: "User 2"},
	}
	expectedTotal := 2

	mockRepo.EXPECT().
		GetUsers(ctx, 20, 0).
		Return(expectedUsers, expectedTotal, nil).
		Times(1)

	users, total, err := service.GetUsers(ctx, 20, 0)

	require.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, expectedTotal, total)
}

func TestAdminService_GetUsers_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAdminRepository(ctrl)
	mockPool := &MockPool{}
	mockRedis := NewMockRedisClient(false)

	service := service_admin.NewAdminService(mockRepo, mockPool, mockRedis)

	ctx := context.Background()

	mockRepo.EXPECT().
		GetUsers(ctx, 20, 0).
		Return([]domain.User{}, 0, nil).
		Times(1)

	users, total, err := service.GetUsers(ctx, 20, 0)

	require.NoError(t, err)
	assert.Empty(t, users)
	assert.Equal(t, 0, total)
}

func TestAdminService_GetUsers_InvalidLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAdminRepository(ctrl)
	mockPool := &MockPool{}
	mockRedis := NewMockRedisClient(false)

	service := service_admin.NewAdminService(mockRepo, mockPool, mockRedis)

	ctx := context.Background()

	users, total, err := service.GetUsers(ctx, -5, 0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "limit must be positive")
	assert.Nil(t, users)
	assert.Equal(t, 0, total)
}

func TestAdminService_GetUsers_InvalidOffset(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAdminRepository(ctrl)
	mockPool := &MockPool{}
	mockRedis := NewMockRedisClient(false)

	service := service_admin.NewAdminService(mockRepo, mockPool, mockRedis)

	ctx := context.Background()

	users, total, err := service.GetUsers(ctx, 20, -5)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "offset must be positive")
	assert.Nil(t, users)
	assert.Equal(t, 0, total)
}

func TestAdminService_GetUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAdminRepository(ctrl)
	mockPool := &MockPool{}
	mockRedis := NewMockRedisClient(true)

	service := service_admin.NewAdminService(mockRepo, mockPool, mockRedis)

	ctx := context.Background()
	userID := 1
	expected := domain.User{ID: 1, FullName: "Admin User", Email: "admin@example.com"}

	mockRepo.EXPECT().
		GetUser(ctx, userID).
		Return(expected, nil).
		Times(1)

	user, err := service.GetUser(ctx, userID)

	require.NoError(t, err)
	assert.Equal(t, expected.ID, user.ID)
	assert.Equal(t, expected.FullName, user.FullName)
	assert.Equal(t, expected.Email, user.Email)
}

func TestAdminService_GetUser_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAdminRepository(ctrl)
	mockPool := &MockPool{}
	mockRedis := NewMockRedisClient(true)

	service := service_admin.NewAdminService(mockRepo, mockPool, mockRedis)

	ctx := context.Background()
	userID := 999

	mockRepo.EXPECT().
		GetUser(ctx, userID).
		Return(domain.User{}, errors.New("not found")).
		Times(1)

	user, err := service.GetUser(ctx, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get user from postgres: not found")
	assert.Equal(t, domain.User{}, user)
}

func TestAdminService_DeleteUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAdminRepository(ctrl)
	mockPool := &MockPool{}
	mockRedis := NewMockRedisClient(false)

	service := service_admin.NewAdminService(mockRepo, mockPool, mockRedis)

	ctx := context.Background()
	userID := 1

	mockRepo.EXPECT().
		DeleteUserTx(ctx, gomock.Any(), userID).
		Return(nil).
		Times(1)

	err := service.DeleteUser(ctx, userID)

	require.NoError(t, err)
}

func TestAdminService_DeleteUser_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAdminRepository(ctrl)
	mockPool := &MockPool{}
	mockRedis := NewMockRedisClient(false)

	service := service_admin.NewAdminService(mockRepo, mockPool, mockRedis)

	ctx := context.Background()
	userID := 999

	mockRepo.EXPECT().
		DeleteUserTx(ctx, gomock.Any(), userID).
		Return(errors.New("not found")).
		Times(1)

	err := service.DeleteUser(ctx, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DeleteUserTx: not found")
}

func TestAdminService_UpdateUserRole_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAdminRepository(ctrl)
	mockPool := &MockPool{}
	mockRedis := NewMockRedisClient(false)

	service := service_admin.NewAdminService(mockRepo, mockPool, mockRedis)

	ctx := context.Background()
	userID := 1
	expected := domain.User{ID: 1, IsAdmin: true}

	mockRepo.EXPECT().
		UpdateUserRoleTx(ctx, gomock.Any(), userID, true).
		Return(expected, nil).
		Times(1)

	user, err := service.UpdateUserRole(ctx, userID, true)

	require.NoError(t, err)
	assert.True(t, user.IsAdmin)
}

func TestAdminService_UpdateUserRole_RemoveAdmin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAdminRepository(ctrl)
	mockPool := &MockPool{}
	mockRedis := NewMockRedisClient(false)

	service := service_admin.NewAdminService(mockRepo, mockPool, mockRedis)

	ctx := context.Background()
	userID := 1
	expected := domain.User{ID: 1, IsAdmin: false}

	mockRepo.EXPECT().
		UpdateUserRoleTx(ctx, gomock.Any(), userID, false).
		Return(expected, nil).
		Times(1)

	user, err := service.UpdateUserRole(ctx, userID, false)

	require.NoError(t, err)
	assert.False(t, user.IsAdmin)
}

func TestAdminService_UpdateUserRole_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAdminRepository(ctrl)
	mockPool := &MockPool{}
	mockRedis := NewMockRedisClient(false)

	service := service_admin.NewAdminService(mockRepo, mockPool, mockRedis)

	ctx := context.Background()
	userID := 999

	mockRepo.EXPECT().
		UpdateUserRoleTx(ctx, gomock.Any(), userID, true).
		Return(domain.User{}, errors.New("not found")).
		Times(1)

	user, err := service.UpdateUserRole(ctx, userID, true)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update user role: not found")
	assert.Equal(t, domain.User{}, user)
}

func TestAdminService_GetMetrics_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAdminRepository(ctrl)
	mockPool := &MockPool{}
	mockRedis := NewMockRedisClient(true)

	service := service_admin.NewAdminService(mockRepo, mockPool, mockRedis)

	ctx := context.Background()
	expected := service_admin.Metrics{
		TotalUsers:        10,
		TotalTransactions: 100,
		TotalBalance:      15000.50,
	}

	mockRepo.EXPECT().
		GetMetrics(ctx).
		Return(expected, nil).
		Times(1)

	metrics, err := service.GetMetrics(ctx)

	require.NoError(t, err)
	assert.Equal(t, 10, metrics.TotalUsers)
	assert.Equal(t, 100, metrics.TotalTransactions)
	assert.Equal(t, 15000.50, metrics.TotalBalance)
}

func TestAdminService_GetMetrics_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAdminRepository(ctrl)
	mockPool := &MockPool{}
	mockRedis := NewMockRedisClient(true)

	service := service_admin.NewAdminService(mockRepo, mockPool, mockRedis)

	ctx := context.Background()

	mockRepo.EXPECT().
		GetMetrics(ctx).
		Return(service_admin.Metrics{}, errors.New("database error")).
		Times(1)

	metrics, err := service.GetMetrics(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get metrics from postgres: database error")
	assert.Equal(t, service_admin.Metrics{}, metrics)
}
