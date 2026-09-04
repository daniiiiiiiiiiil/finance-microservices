package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/repository/postgres/pool"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/features/finance/repository/postgres"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/pkg/logger"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockFinanceRepository struct {
	mock.Mock
}

func (m *MockFinanceRepository) CreateTransactionTx(ctx context.Context, tx pool.Tx, transaction domain.Finance) (domain.Finance, error) {
	args := m.Called(ctx, tx, transaction)
	if args.Get(0) == nil {
		return domain.Finance{}, args.Error(1)
	}
	return args.Get(0).(domain.Finance), args.Error(1)
}

func (m *MockFinanceRepository) GetTransaction(ctx context.Context, id int) (domain.Finance, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.Finance), args.Error(1)
}

func (m *MockFinanceRepository) GetTransactions(ctx context.Context, userID int, transactionType, category *string, from, to *time.Time, limit, offset int) ([]domain.Finance, error) {
	args := m.Called(ctx, userID, transactionType, category, from, to, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Finance), args.Error(1)
}

func (m *MockFinanceRepository) UpdateTransaction(ctx context.Context, transaction domain.Finance) (domain.Finance, error) {
	args := m.Called(ctx, transaction)
	return args.Get(0).(domain.Finance), args.Error(1)
}

func (m *MockFinanceRepository) DeleteTransaction(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockFinanceRepository) GetCategories(ctx context.Context, userID int) ([]string, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockFinanceRepository) GetDashboard(ctx context.Context, userID int) (domain.Dashboard, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(domain.Dashboard), args.Error(1)
}

func (m *MockFinanceRepository) DeleteUserTransactions(ctx context.Context, userID int) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *MockFinanceRepository) GetMetrics(ctx context.Context) (postgres.Metrics, error) {
	args := m.Called(ctx)
	return args.Get(0).(postgres.Metrics), args.Error(1)
}

func (m *MockFinanceRepository) GetTransactionsCount(ctx context.Context, userID int, transactionType, category *string, from, to *time.Time) (int, error) {
	args := m.Called(ctx, userID, transactionType, category, from, to)
	return args.Int(0), args.Error(1)
}

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

type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) Publish(ctx context.Context, eventType string, data interface{}) error {
	args := m.Called(ctx, eventType, data)
	return args.Error(0)
}

type MockPool struct {
	mock.Mock
}

func (m *MockPool) Query(ctx context.Context, sql string, args ...any) (pool.Rows, error) {
	callArgs := m.Called(ctx, sql, args)
	if callArgs.Get(0) == nil {
		return nil, callArgs.Error(1)
	}
	return callArgs.Get(0).(pool.Rows), callArgs.Error(1)
}

func (m *MockPool) QueryRow(ctx context.Context, sql string, args ...any) pool.Row {
	callArgs := m.Called(ctx, sql, args)
	if callArgs.Get(0) == nil {
		return nil
	}
	return callArgs.Get(0).(pool.Row)
}

func (m *MockPool) Exec(ctx context.Context, sql string, args ...any) (pool.CommandTag, error) {
	callArgs := m.Called(ctx, sql, args)
	if callArgs.Get(0) == nil {
		return nil, callArgs.Error(1)
	}
	return callArgs.Get(0).(pool.CommandTag), callArgs.Error(1)
}

func (m *MockPool) Begin(ctx context.Context) (pool.Tx, error) {
	callArgs := m.Called(ctx)
	if callArgs.Get(0) == nil {
		return nil, callArgs.Error(1)
	}
	return callArgs.Get(0).(pool.Tx), callArgs.Error(1)
}

func (m *MockPool) Close() {
	m.Called()
}

func (m *MockPool) OpTimeout() time.Duration {
	callArgs := m.Called()
	return callArgs.Get(0).(time.Duration)
}

type MockOutboxRepo struct {
	mock.Mock
}

func (m *MockOutboxRepo) SaveTx(ctx context.Context, tx pool.Tx, event domain.OutboxEvent) error {
	args := m.Called(ctx, tx, event)
	return args.Error(0)
}

func (m *MockOutboxRepo) Save(ctx context.Context, event domain.OutboxEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockOutboxRepo) GetPending(ctx context.Context, limit int) ([]domain.OutboxEvent, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.OutboxEvent), args.Error(1)
}

func (m *MockOutboxRepo) MarkProcessed(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOutboxRepo) MarkFailed(ctx context.Context, id string, errMsg string) error {
	args := m.Called(ctx, id, errMsg)
	return args.Error(0)
}

type testSuite struct {
	service       *FinanceService
	mockRepo      *MockFinanceRepository
	mockPool      *MockPool
	mockRedis     *MockRedis
	mockPublisher *MockPublisher
	mockOutbox    *MockOutboxRepo
}

func setup() *testSuite {
	mockRepo := new(MockFinanceRepository)
	mockPool := new(MockPool)
	mockRedis := new(MockRedis)
	mockPublisher := new(MockPublisher)
	mockOutbox := new(MockOutboxRepo)

	loggerInstance := &logger.Logger{Logger: zap.NewNop()}

	service := &FinanceService{
		repo:           mockRepo,
		pool:           mockPool,
		redis:          mockRedis,
		eventPublisher: mockPublisher,
		outboxRepo:     mockOutbox,
		logger:         loggerInstance,
	}

	return &testSuite{
		service:       service,
		mockRepo:      mockRepo,
		mockPool:      mockPool,
		mockRedis:     mockRedis,
		mockPublisher: mockPublisher,
		mockOutbox:    mockOutbox,
	}
}

type MockTx struct {
	mock.Mock
}

func (m *MockTx) Exec(ctx context.Context, sql string, args ...any) (pool.CommandTag, error) {
	callArgs := m.Called(ctx, sql, args)
	if callArgs.Get(0) == nil {
		return nil, callArgs.Error(1)
	}
	return callArgs.Get(0).(pool.CommandTag), callArgs.Error(1)
}

func (m *MockTx) Query(ctx context.Context, sql string, args ...any) (pool.Rows, error) {
	callArgs := m.Called(ctx, sql, args)
	if callArgs.Get(0) == nil {
		return nil, callArgs.Error(1)
	}
	return callArgs.Get(0).(pool.Rows), callArgs.Error(1)
}

func (m *MockTx) QueryRow(ctx context.Context, sql string, args ...any) pool.Row {
	callArgs := m.Called(ctx, sql, args)
	if callArgs.Get(0) == nil {
		return nil
	}
	return callArgs.Get(0).(pool.Row)
}

func (m *MockTx) Commit(ctx context.Context) error {
	callArgs := m.Called(ctx)
	return callArgs.Error(0)
}

func (m *MockTx) Rollback(ctx context.Context) error {
	callArgs := m.Called(ctx)
	return callArgs.Error(0)
}

func TestCreateTransaction_Success(t *testing.T) {
	s := setup()
	ctx := context.Background()

	transaction := domain.Finance{
		TypeTransaction: "income",
		Amount:          1000.00,
		Category:        "salary",
		UserID:          1,
		CreatedAt:       time.Now(),
	}

	expected := domain.Finance{
		ID:              1,
		Version:         1,
		TypeTransaction: "income",
		Amount:          1000.00,
		Category:        "salary",
		UserID:          1,
		CreatedAt:       transaction.CreatedAt,
	}

	mockTx := new(MockTx)
	s.mockPool.On("Begin", ctx).Return(mockTx, nil)
	mockTx.On("Rollback", ctx).Return(nil)
	mockTx.On("Commit", ctx).Return(nil)
	s.mockRepo.On("CreateTransactionTx", ctx, mockTx, transaction).Return(expected, nil)

	s.mockOutbox.On("SaveTx", ctx, mockTx, mock.Anything).Return(nil)

	s.mockRedis.On("Delete", ctx, "dashboard:1").Return(nil)
	s.mockRedis.On("Delete", ctx, "categories:1").Return(nil)

	result, err := s.service.CreateTransaction(ctx, transaction)

	assert.NoError(t, err)
	assert.Equal(t, 1, result.ID)
	assert.Equal(t, 1000.00, result.Amount)
	s.mockRepo.AssertExpectations(t)
	s.mockPool.AssertExpectations(t)
	s.mockOutbox.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

func TestCreateTransaction_ValidationError(t *testing.T) {
	s := setup()
	ctx := context.Background()

	transaction := domain.Finance{
		TypeTransaction: "invalid",
		Amount:          -100.00,
		Category:        "",
		UserID:          1,
	}

	mockTx := new(MockTx)
	s.mockPool.On("Begin", ctx).Return(mockTx, nil)
	mockTx.On("Rollback", ctx).Return(nil)

	_, err := s.service.CreateTransaction(ctx, transaction)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestGetTransaction_Success(t *testing.T) {
	s := setup()
	ctx := context.Background()

	expected := domain.Finance{
		ID:              1,
		TypeTransaction: "income",
		Amount:          1000.00,
		Category:        "salary",
		UserID:          1,
	}

	s.mockRepo.On("GetTransaction", ctx, 1).Return(expected, nil)

	result, err := s.service.GetTransaction(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, 1, result.ID)
	assert.Equal(t, 1000.00, result.Amount)
	s.mockRepo.AssertExpectations(t)
}

func TestGetTransaction_NotFound(t *testing.T) {
	s := setup()
	ctx := context.Background()

	s.mockRepo.On("GetTransaction", ctx, 999).Return(domain.Finance{}, errors.New("not found"))

	_, err := s.service.GetTransaction(ctx, 999)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get transaction")
}

func TestUpdateTransaction_Success(t *testing.T) {
	s := setup()
	ctx := context.Background()

	existing := domain.Finance{
		ID:              1,
		Version:         1,
		TypeTransaction: "income",
		Amount:          100.00,
		Category:        "salary",
		UserID:          1,
		CreatedAt:       time.Now(),
	}

	updated := domain.Finance{
		ID:              1,
		TypeTransaction: "income",
		Amount:          200.00,
		Category:        "salary",
	}

	s.mockRepo.On("GetTransaction", ctx, 1).Return(existing, nil)
	s.mockRepo.On("UpdateTransaction", ctx, mock.Anything).Return(domain.Finance{
		ID:              1,
		Version:         2,
		TypeTransaction: "income",
		Amount:          200.00,
		Category:        "salary",
		UserID:          1,
		CreatedAt:       existing.CreatedAt,
	}, nil)

	s.mockOutbox.On("Save", ctx, mock.Anything).Return(nil)

	s.mockRedis.On("Delete", ctx, "dashboard:1").Return(nil)
	s.mockRedis.On("Delete", ctx, "categories:1").Return(nil)

	result, err := s.service.UpdateTransaction(ctx, updated)

	assert.NoError(t, err)
	assert.Equal(t, 2, result.Version)
	assert.Equal(t, 200.00, result.Amount)
	s.mockRepo.AssertExpectations(t)
	s.mockOutbox.AssertExpectations(t)
}

func TestUpdateTransaction_NotFound(t *testing.T) {
	s := setup()
	ctx := context.Background()

	transaction := domain.Finance{ID: 999}

	s.mockRepo.On("GetTransaction", ctx, 999).Return(domain.Finance{}, errors.New("not found"))

	_, err := s.service.UpdateTransaction(ctx, transaction)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get existing transaction")
}

func TestDeleteTransaction_Success(t *testing.T) {
	s := setup()
	ctx := context.Background()

	existing := domain.Finance{
		ID:     1,
		UserID: 1,
	}

	s.mockRepo.On("GetTransaction", ctx, 1).Return(existing, nil)

	s.mockRepo.On("DeleteTransaction", ctx, 1).Return(nil)

	s.mockOutbox.On("Save", ctx, mock.Anything).Return(nil)

	s.mockRedis.On("Delete", ctx, "dashboard:1").Return(nil)
	s.mockRedis.On("Delete", ctx, "categories:1").Return(nil)

	err := s.service.DeleteTransaction(ctx, 1)

	assert.NoError(t, err)
	s.mockRepo.AssertExpectations(t)
	s.mockOutbox.AssertExpectations(t)
	s.mockRedis.AssertExpectations(t)
}

func TestDeleteTransaction_NotFound(t *testing.T) {
	s := setup()
	ctx := context.Background()

	s.mockRepo.On("GetTransaction", ctx, 999).Return(domain.Finance{}, errors.New("not found"))

	err := s.service.DeleteTransaction(ctx, 999)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get transaction")
}

func TestGetDashboard_Success(t *testing.T) {
	s := setup()
	ctx := context.Background()

	expected := domain.Dashboard{
		TotalBalance:    15000.50,
		MonthlyIncome:   5000.00,
		MonthlyExpenses: 2000.00,
		SavingsRate:     60.0,
	}

	s.mockRedis.On("Get", ctx, "dashboard:1", mock.Anything).Return(redis.Nil)
	s.mockRepo.On("GetDashboard", ctx, 1).Return(expected, nil)
	s.mockRedis.On("Set", ctx, "dashboard:1", expected, 10*time.Minute).Return(nil)

	result, err := s.service.GetDashboard(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, 15000.50, result.TotalBalance)
	assert.Equal(t, 60.0, result.SavingsRate)
	s.mockRepo.AssertExpectations(t)
}

func TestGetDashboard_FromCache(t *testing.T) {
	s := setup()
	ctx := context.Background()

	expected := domain.Dashboard{
		TotalBalance: 10000.00,
	}

	s.mockRedis.On("Get", ctx, "dashboard:1", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		dest := args.Get(2).(*domain.Dashboard)
		*dest = expected
	})

	result, err := s.service.GetDashboard(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, 10000.00, result.TotalBalance)
	s.mockRepo.AssertNotCalled(t, "GetDashboard", mock.Anything, mock.Anything)
}

func TestGetCategories_Success(t *testing.T) {
	s := setup()
	ctx := context.Background()

	expected := []string{"food", "salary", "transport"}

	s.mockRedis.On("Get", ctx, "categories:1", mock.Anything).Return(errors.New("not found"))
	s.mockRepo.On("GetCategories", ctx, 1).Return(expected, nil)
	s.mockRedis.On("Set", ctx, "categories:1", expected, 24*time.Hour).Return(nil)

	result, err := s.service.GetCategories(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	s.mockRepo.AssertExpectations(t)
}

func TestDeleteUserTransactions_Success(t *testing.T) {
	s := setup()
	ctx := context.Background()

	s.mockRedis.On("Exists", ctx, "deleted:user_transactions:1").Return(int64(0), nil)
	s.mockRepo.On("DeleteUserTransactions", ctx, 1).Return(5, nil)
	s.mockRedis.On("Set", ctx, "deleted:user_transactions:1", true, 24*time.Hour).Return(nil)
	s.mockRedis.On("Delete", ctx, "dashboard:1").Return(nil)
	s.mockRedis.On("Delete", ctx, "categories:1").Return(nil)
	s.mockPublisher.On("Publish", ctx, "user.transactions.deleted", mock.Anything).Return(nil)

	count, err := s.service.DeleteUserTransactions(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, 5, count)
	s.mockRepo.AssertExpectations(t)
}

func TestDeleteUserTransactions_AlreadyDeleted(t *testing.T) {
	s := setup()
	ctx := context.Background()

	s.mockRedis.On("Exists", ctx, "deleted:user_transactions:1").Return(int64(1), nil)

	count, err := s.service.DeleteUserTransactions(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, 0, count)
	s.mockRepo.AssertNotCalled(t, "DeleteUserTransactions", mock.Anything, mock.Anything)
}

func TestGetMetrics_Success(t *testing.T) {
	s := setup()
	ctx := context.Background()

	expected := postgres.Metrics{
		TotalTransactions: 100,
		TotalBalance:      5000.50,
	}

	s.mockRedis.On("Get", ctx, "finance:metrics", mock.Anything).Return(errors.New("not found"))
	s.mockRepo.On("GetMetrics", ctx).Return(expected, nil)
	s.mockRedis.On("Set", ctx, "finance:metrics", expected, time.Minute).Return(nil)

	result, err := s.service.GetMetrics(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 100, result.TotalTransactions)
	assert.Equal(t, 5000.50, result.TotalBalance)
	s.mockRepo.AssertExpectations(t)
}

func TestGetTransactions_Success(t *testing.T) {
	s := setup()
	ctx := context.Background()

	expected := []domain.Finance{
		{ID: 1, Amount: 100.00, UserID: 1},
		{ID: 2, Amount: 200.00, UserID: 1},
	}

	var transactionType *string
	var category *string
	var from *time.Time
	var to *time.Time

	s.mockRepo.On("GetTransactions", ctx, 1, transactionType, category, from, to, 20, 0).Return(expected, nil)

	result, err := s.service.GetTransactions(ctx, 1, transactionType, category, from, to, 20, 0)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 1, result[0].ID)
}

func TestGetTransactions_InvalidDateRange(t *testing.T) {
	s := setup()
	ctx := context.Background()

	from := time.Now()
	to := time.Now().Add(-24 * time.Hour)

	_, err := s.service.GetTransactions(ctx, 1, nil, nil, &from, &to, 20, 0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "from date must be before to date")
}

func TestCreateTransaction_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		transaction domain.Finance
		expectError bool
	}{
		{
			name: "valid income transaction",
			transaction: domain.Finance{
				TypeTransaction: "income",
				Amount:          1000.00,
				Category:        "salary",
				UserID:          1,
			},
			expectError: false,
		},
		{
			name: "valid expense transaction",
			transaction: domain.Finance{
				TypeTransaction: "expense",
				Amount:          500.00,
				Category:        "food",
				UserID:          1,
			},
			expectError: false,
		},
		{
			name: "invalid type",
			transaction: domain.Finance{
				TypeTransaction: "invalid",
				Amount:          1000.00,
				Category:        "salary",
				UserID:          1,
			},
			expectError: true,
		},
		{
			name: "negative amount",
			transaction: domain.Finance{
				TypeTransaction: "income",
				Amount:          -100.00,
				Category:        "salary",
				UserID:          1,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := setup()
			ctx := context.Background()

			mockTx := new(MockTx)
			s.mockPool.On("Begin", ctx).Return(mockTx, nil)
			mockTx.On("Rollback", ctx).Return(nil)

			if !tt.expectError {
				mockTx.On("Commit", ctx).Return(nil)
				s.mockRepo.On("CreateTransactionTx", ctx, mockTx, tt.transaction).Return(domain.Finance{ID: 1}, nil)

				s.mockOutbox.On("SaveTx", ctx, mockTx, mock.Anything).Return(nil)

				s.mockRedis.On("Delete", mock.Anything, mock.Anything).Return(nil)
				s.mockPublisher.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			}

			_, err := s.service.CreateTransaction(ctx, tt.transaction)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetDashboard_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		userID      int
		cached      bool
		expectError bool
	}{
		{
			name:        "from cache",
			userID:      1,
			cached:      true,
			expectError: false,
		},
		{
			name:        "from database",
			userID:      1,
			cached:      false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := setup()
			ctx := context.Background()

			dashboardKey := fmt.Sprintf("dashboard:%d", tt.userID)

			if tt.cached {
				s.mockRedis.On("Get", ctx, dashboardKey, mock.Anything).Return(nil)
			} else {
				s.mockRedis.On("Get", ctx, dashboardKey, mock.Anything).Return(redis.Nil)
				s.mockRepo.On("GetDashboard", ctx, tt.userID).Return(domain.Dashboard{}, nil)
				s.mockRedis.On("Set", ctx, dashboardKey, mock.Anything, 10*time.Minute).Return(nil)
			}

			_, err := s.service.GetDashboard(ctx, tt.userID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
