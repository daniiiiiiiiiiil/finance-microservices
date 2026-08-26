package test

import (
	"backend/internal/core/domain"
	"backend/internal/core/repository/postgres/pool"
	service_finance "backend/internal/features/finance/service"
	"backend/internal/features/finance/service/mocks"
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

func TestFinanceService_GetDashboard_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockFinanceRepository(ctrl)
	mockPool := &MockPool{}
	mockRedis := NewMockRedisClient(true)

	service := service_finance.NewFinanceService(mockRepo, mockPool, mockRedis)

	ctx := context.Background()
	userID := 1
	expected := domain.Dashboard{
		TotalBalance:  15000.50,
		MonthlyIncome: 5000.00,
	}

	mockRepo.EXPECT().
		GetDashboard(ctx, userID).
		Return(expected, nil).
		Times(1)

	dashboard, err := service.GetDashboard(ctx, userID)

	require.NoError(t, err)
	assert.Equal(t, expected.TotalBalance, dashboard.TotalBalance)
}

func TestFinanceService_GetDashboard_FromRedis(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockFinanceRepository(ctrl)
	mockPool := &MockPool{}
	mockRedis := NewMockRedisClient(true)

	service := service_finance.NewFinanceService(mockRepo, mockPool, mockRedis)

	ctx := context.Background()
	userID := 1

	mockRepo.EXPECT().
		GetDashboard(ctx, userID).
		Return(domain.Dashboard{TotalBalance: 1000}, nil).
		Times(1)

	dashboard, err := service.GetDashboard(ctx, userID)

	require.NoError(t, err)
	assert.Equal(t, 1000.00, dashboard.TotalBalance)
}

func TestFinanceService_GetDashboard_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockFinanceRepository(ctrl)
	mockPool := &MockPool{}
	mockRedis := NewMockRedisClient(true)

	service := service_finance.NewFinanceService(mockRepo, mockPool, mockRedis)

	ctx := context.Background()
	userID := 1

	mockRepo.EXPECT().
		GetDashboard(ctx, userID).
		Return(domain.Dashboard{}, errors.New("database error")).
		Times(1)

	dashboard, err := service.GetDashboard(ctx, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to get dashboard")
	assert.Equal(t, domain.Dashboard{}, dashboard)
}

func TestFinanceService_GetCategories_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockFinanceRepository(ctrl)
	mockPool := &MockPool{}
	mockRedis := NewMockRedisClient(true)

	service := service_finance.NewFinanceService(mockRepo, mockPool, mockRedis)

	ctx := context.Background()
	userID := 1
	expected := []string{"food", "salary", "transport"}

	mockRepo.EXPECT().
		GetCategories(ctx, userID).
		Return(expected, nil).
		Times(1)

	categories, err := service.GetCategories(ctx, userID)

	require.NoError(t, err)
	assert.Equal(t, expected, categories)
}

func TestFinanceService_GetCategories_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockFinanceRepository(ctrl)
	mockPool := &MockPool{}
	mockRedis := NewMockRedisClient(true)

	service := service_finance.NewFinanceService(mockRepo, mockPool, mockRedis)

	ctx := context.Background()
	userID := 1

	mockRepo.EXPECT().
		GetCategories(ctx, userID).
		Return(nil, errors.New("database error")).
		Times(1)

	categories, err := service.GetCategories(ctx, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get categories")
	assert.Nil(t, categories)
}

func TestFinanceService_CreateTransaction_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockFinanceRepository(ctrl)
	mockPool := &MockPool{}
	mockRedis := NewMockRedisClient(false)

	service := service_finance.NewFinanceService(mockRepo, mockPool, mockRedis)

	ctx := context.Background()
	transaction := domain.Finance{
		TypeTransaction: "income",
		Amount:          1000.00,
		Category:        "salary",
		UserID:          1,
	}
	expected := domain.Finance{ID: 1, Amount: 1000.00}

	mockRepo.EXPECT().
		CreateTransactionTx(ctx, gomock.Any(), transaction).
		Return(expected, nil).
		Times(1)

	result, err := service.CreateTransaction(ctx, transaction)

	require.NoError(t, err)
	assert.Equal(t, expected.ID, result.ID)
}

func TestFinanceService_CreateTransaction_InvalidData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockFinanceRepository(ctrl)
	mockPool := &MockPool{}
	mockRedis := NewMockRedisClient(false)

	service := service_finance.NewFinanceService(mockRepo, mockPool, mockRedis)

	ctx := context.Background()
	transaction := domain.Finance{
		TypeTransaction: "invalid",
		Amount:          -100.00,
		Category:        "",
		UserID:          1,
	}

	result, err := service.CreateTransaction(ctx, transaction)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
	assert.Equal(t, domain.Finance{}, result)
}

func TestFinanceService_GetTransaction_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockFinanceRepository(ctrl)
	mockPool := &MockPool{}
	mockRedis := NewMockRedisClient(false)

	service := service_finance.NewFinanceService(mockRepo, mockPool, mockRedis)

	ctx := context.Background()
	transactionID := 1
	expected := domain.Finance{ID: 1, Amount: 1000.00}

	mockRepo.EXPECT().
		GetTransaction(ctx, transactionID).
		Return(expected, nil).
		Times(1)

	result, err := service.GetTransaction(ctx, transactionID)

	require.NoError(t, err)
	assert.Equal(t, expected.ID, result.ID)
}

func TestFinanceService_GetTransaction_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockFinanceRepository(ctrl)
	mockPool := &MockPool{}
	mockRedis := NewMockRedisClient(false)

	service := service_finance.NewFinanceService(mockRepo, mockPool, mockRedis)

	ctx := context.Background()
	transactionID := 999

	mockRepo.EXPECT().
		GetTransaction(ctx, transactionID).
		Return(domain.Finance{}, errors.New("not found")).
		Times(1)

	result, err := service.GetTransaction(ctx, transactionID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get transaction: not found")
	assert.Equal(t, domain.Finance{}, result)
}

func TestFinanceService_GetTransactions_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockFinanceRepository(ctrl)
	mockPool := &MockPool{}
	mockRedis := NewMockRedisClient(false)

	service := service_finance.NewFinanceService(mockRepo, mockPool, mockRedis)

	ctx := context.Background()
	userID := 1
	expected := []domain.Finance{
		{ID: 1, Amount: 1000.00},
		{ID: 2, Amount: 500.00},
	}

	mockRepo.EXPECT().
		GetTransactions(ctx, userID, nil, nil, nil, nil, 20, 0).
		Return(expected, nil).
		Times(1)

	result, err := service.GetTransactions(ctx, userID, nil, nil, nil, nil, 20, 0)

	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestFinanceService_GetTransactions_InvalidDateRange(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockFinanceRepository(ctrl)
	mockPool := &MockPool{}
	mockRedis := NewMockRedisClient(false)

	service := service_finance.NewFinanceService(mockRepo, mockPool, mockRedis)

	ctx := context.Background()
	userID := 1
	from := time.Now()
	to := time.Now().Add(-24 * time.Hour)

	result, err := service.GetTransactions(ctx, userID, nil, nil, &from, &to, 20, 0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "from date must be before to date")
	assert.Nil(t, result)
}

func TestFinanceService_UpdateTransaction_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockFinanceRepository(ctrl)
	mockPool := &MockPool{}
	mockRedis := NewMockRedisClient(false)

	service := service_finance.NewFinanceService(mockRepo, mockPool, mockRedis)

	ctx := context.Background()
	transactionID := 1
	existing := domain.Finance{ID: 1, Version: 1, Amount: 100.00, TypeTransaction: "income", Category: "salary"}
	updated := domain.Finance{ID: 1, Amount: 200.00, Version: 1, TypeTransaction: "income", Category: "salary"}

	mockRepo.EXPECT().
		GetTransaction(ctx, transactionID).
		Return(existing, nil).
		Times(1)

	mockRepo.EXPECT().
		UpdateTransaction(ctx, gomock.Any()).
		Return(domain.Finance{ID: 1, Version: 2, Amount: 200.00, TypeTransaction: "income", Category: "salary"}, nil).
		Times(1)

	result, err := service.UpdateTransaction(ctx, updated)

	require.NoError(t, err)
	assert.Equal(t, 2, result.Version)
}

func TestFinanceService_UpdateTransaction_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockFinanceRepository(ctrl)
	mockPool := &MockPool{}
	mockRedis := NewMockRedisClient(false)

	service := service_finance.NewFinanceService(mockRepo, mockPool, mockRedis)

	ctx := context.Background()
	transactionID := 999
	transaction := domain.Finance{ID: 999}

	mockRepo.EXPECT().
		GetTransaction(ctx, transactionID).
		Return(domain.Finance{}, errors.New("not found")).
		Times(1)

	result, err := service.UpdateTransaction(ctx, transaction)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get existing transaction: not found")
	assert.Equal(t, domain.Finance{}, result)
}

func TestFinanceService_DeleteTransaction_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockFinanceRepository(ctrl)
	mockPool := &MockPool{}
	mockRedis := NewMockRedisClient(false)

	service := service_finance.NewFinanceService(mockRepo, mockPool, mockRedis)

	ctx := context.Background()
	transactionID := 1

	mockRepo.EXPECT().
		DeleteTransaction(ctx, transactionID).
		Return(nil).
		Times(1)

	err := service.DeleteTransaction(ctx, transactionID)

	require.NoError(t, err)
}

func TestFinanceService_DeleteTransaction_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockFinanceRepository(ctrl)
	mockPool := &MockPool{}
	mockRedis := NewMockRedisClient(false)

	service := service_finance.NewFinanceService(mockRepo, mockPool, mockRedis)

	ctx := context.Background()
	transactionID := 999

	mockRepo.EXPECT().
		DeleteTransaction(ctx, transactionID).
		Return(errors.New("not found")).
		Times(1)

	err := service.DeleteTransaction(ctx, transactionID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete transaction: not found")
}
