package service_user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/repository/postgres/pool"
	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockUsersRepository struct {
	mock.Mock
}

func (m *MockUsersRepository) GetUser(ctx context.Context, id int) (domain.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *MockUsersRepository) DeleteUserTx(ctx context.Context, tx pool.Tx, id int) error {
	args := m.Called(ctx, tx, id)
	return args.Error(0)
}

func (m *MockUsersRepository) PatchUser(ctx context.Context, id int, patch domain.User) (domain.User, error) {
	args := m.Called(ctx, id, patch)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *MockUsersRepository) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *MockUsersRepository) CreateUser(ctx context.Context, user domain.User) (int, error) {
	args := m.Called(ctx, user)
	return args.Int(0), args.Error(1)
}

func (m *MockUsersRepository) ListUsers(ctx context.Context, limit, offset int) ([]domain.User, int, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]domain.User), args.Int(1), args.Error(2)
}

func (m *MockUsersRepository) UpdateRole(ctx context.Context, id int, isAdmin bool) (domain.User, error) {
	args := m.Called(ctx, id, isAdmin)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *MockUsersRepository) GetTotalUsers(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *MockUsersRepository) AdminExists(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}

func (m *MockUsersRepository) UpdateStatusTx(ctx context.Context, tx pool.Tx, id int, status string) error {
	args := m.Called(ctx, tx, id, status)
	return args.Error(0)
}

type MockPool struct {
	mock.Mock
}

func (m *MockPool) Begin(ctx context.Context) (pool.Tx, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(pool.Tx), args.Error(1)
}

func (m *MockPool) OpTimeout() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
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

type MockUserCache struct {
	mock.Mock
}

func (m *MockUserCache) GetUser(ctx context.Context, userID int) (domain.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(domain.User), args.Error(1)
}

func (m *MockUserCache) SetUser(ctx context.Context, user domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserCache) DeleteUser(ctx context.Context, userID int) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserCache) InvalidateUser(ctx context.Context, userID int) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

type MockUsersListCache struct {
	mock.Mock
}

func (m *MockUsersListCache) GetUsersList(ctx context.Context, limit, offset int) ([]domain.User, bool) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]domain.User), args.Bool(1)
}

func (m *MockUsersListCache) SetUsersList(ctx context.Context, users []domain.User, limit, offset int) error {
	args := m.Called(ctx, users, limit, offset)
	return args.Error(0)
}

func (m *MockUsersListCache) InvalidateAllUsersList(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) Publish(ctx context.Context, eventType string, data interface{}) error {
	args := m.Called(ctx, eventType, data)
	return args.Error(0)
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
	service       *UsersService
	mockRepo      *MockUsersRepository
	mockPool      *MockPool
	mockUserCache *MockUserCache
	mockListCache *MockUsersListCache
	mockPublisher *MockPublisher
	mockRedis     *MockRedis
	mockOutbox    *MockOutboxRepo
}

func setup() *testSuite {
	mockRepo := new(MockUsersRepository)
	mockPool := new(MockPool)
	mockUserCache := new(MockUserCache)
	mockListCache := new(MockUsersListCache)
	mockPublisher := new(MockPublisher)
	mockRedis := new(MockRedis)
	mockOutbox := new(MockOutboxRepo)

	loggerInstance := &logger.Logger{Logger: zap.NewNop()}

	service := &UsersService{
		userRepository: mockRepo,
		pool:           mockPool,
		userCache:      mockUserCache,
		usersListCache: mockListCache,
		eventPublisher: mockPublisher,
		outboxRepo:     mockOutbox,
		logger:         loggerInstance,
		redis:          mockRedis,
	}

	return &testSuite{
		service:       service,
		mockRepo:      mockRepo,
		mockPool:      mockPool,
		mockUserCache: mockUserCache,
		mockListCache: mockListCache,
		mockPublisher: mockPublisher,
		mockRedis:     mockRedis,
		mockOutbox:    mockOutbox,
	}
}

func TestCreateProfile_Success(t *testing.T) {
	s := setup()
	ctx := context.Background()

	req := &CreateProfileRequest{
		Email:        "john@example.com",
		FullName:     "John Doe",
		PhoneNumber:  nil,
		IsAdmin:      false,
		PasswordHash: "hashed_password",
	}

	s.mockRepo.On("GetUserByEmail", ctx, req.Email).Return(domain.User{}, errors.New("not found"))
	s.mockRepo.On("CreateUser", ctx, mock.Anything).Return(1, nil)
	s.mockRepo.On("GetUser", ctx, 1).Return(domain.User{
		ID:           1,
		FullName:     "John Doe",
		Email:        "john@example.com",
		PasswordHash: "hashed_password",
		IsAdmin:      false,
		Status:       "active",
	}, nil)
	s.mockPublisher.On("Publish", ctx, "user.created", mock.Anything).Return(nil)

	user, err := s.service.CreateProfile(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, 1, user.ID)
	assert.Equal(t, "john@example.com", user.Email)
	s.mockRepo.AssertExpectations(t)
	s.mockOutbox.AssertExpectations(t)
}

func TestCreateProfile_AlreadyExists(t *testing.T) {
	s := setup()
	ctx := context.Background()

	req := &CreateProfileRequest{
		Email:        "john@example.com",
		FullName:     "John Doe",
		PasswordHash: "hashed_password",
	}

	s.mockRepo.On("GetUserByEmail", ctx, req.Email).Return(domain.User{ID: 1}, nil)

	_, err := s.service.CreateProfile(ctx, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user already exists")
}

func TestGetUser_Success(t *testing.T) {
	s := setup()
	ctx := context.Background()

	s.mockUserCache.On("GetUser", ctx, 1).Return(domain.User{}, errors.New("not found"))
	s.mockRepo.On("GetUser", ctx, 1).Return(domain.User{
		ID:       1,
		FullName: "John Doe",
		Email:    "john@example.com",
	}, nil)
	s.mockUserCache.On("SetUser", ctx, mock.Anything).Return(nil)

	user, err := s.service.GetUser(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, 1, user.ID)
	assert.Equal(t, "John Doe", user.FullName)
	s.mockRepo.AssertExpectations(t)
}

func TestGetUser_FromCache(t *testing.T) {
	s := setup()
	ctx := context.Background()

	expected := domain.User{
		ID:       1,
		FullName: "John Doe",
		Email:    "john@example.com",
	}

	s.mockUserCache.On("GetUser", ctx, 1).Return(expected, nil)

	user, err := s.service.GetUser(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, expected, user)
	s.mockRepo.AssertNotCalled(t, "GetUser", mock.Anything, mock.Anything)
}

func TestGetUser_NotFound(t *testing.T) {
	s := setup()
	ctx := context.Background()

	s.mockUserCache.On("GetUser", ctx, 999).Return(domain.User{}, errors.New("not found"))
	s.mockRepo.On("GetUser", ctx, 999).Return(domain.User{}, errors.New("not found"))

	_, err := s.service.GetUser(ctx, 999)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "userRepository.GetUser: not found")
}

func TestPatchUser_Success(t *testing.T) {
	s := setup()
	ctx := context.Background()

	existingUser := domain.User{
		ID:       1,
		FullName: "Old Name",
		Email:    "john@example.com",
	}

	patch := domain.UserPatch{
		FullName: domain.Nullable[string]{Value: stringPtr("New Name"), Set: true},
	}

	s.mockRepo.On("GetUser", ctx, 1).Return(existingUser, nil)
	s.mockRepo.On("PatchUser", ctx, 1, mock.Anything).Return(domain.User{
		ID:       1,
		FullName: "New Name",
		Email:    "john@example.com",
	}, nil)
	s.mockUserCache.On("InvalidateUser", ctx, 1).Return(nil)
	s.mockListCache.On("InvalidateAllUsersList", ctx).Return(nil)

	user, err := s.service.PatchUser(ctx, 1, patch)

	assert.NoError(t, err)
	assert.Equal(t, "New Name", user.FullName)
	s.mockRepo.AssertExpectations(t)
}

func TestPatchUser_NotFound(t *testing.T) {
	s := setup()
	ctx := context.Background()

	patch := domain.UserPatch{
		FullName: domain.Nullable[string]{Value: stringPtr("New Name"), Set: true},
	}

	s.mockRepo.On("GetUser", ctx, 999).Return(domain.User{}, errors.New("not found"))

	_, err := s.service.PatchUser(ctx, 999, patch)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "userRepository.GetUser: not found")
}

func TestDeleteUser_Success(t *testing.T) {
	s := setup()
	ctx := context.Background()

	mockTx := new(MockTx)
	s.mockPool.On("Begin", ctx).Return(mockTx, nil)
	mockTx.On("Rollback", ctx).Return(nil)
	mockTx.On("Commit", ctx).Return(nil)
	s.mockRepo.On("DeleteUserTx", ctx, mockTx, 1).Return(nil)
	s.mockUserCache.On("InvalidateUser", ctx, 1).Return(nil)
	s.mockListCache.On("InvalidateAllUsersList", ctx).Return(nil)

	err := s.service.DeleteUser(ctx, 1)

	assert.NoError(t, err)
	s.mockRepo.AssertExpectations(t)
}

func TestMarkDeleting_Success(t *testing.T) {
	s := setup()
	ctx := context.Background()

	mockTx := new(MockTx)
	s.mockPool.On("Begin", ctx).Return(mockTx, nil)
	mockTx.On("Rollback", ctx).Return(nil)
	mockTx.On("Commit", ctx).Return(nil)
	s.mockRepo.On("UpdateStatusTx", ctx, mockTx, 1, "deleting").Return(nil)

	err := s.service.MarkDeleting(ctx, 1)

	assert.NoError(t, err)
	s.mockRepo.AssertExpectations(t)
}

func TestMarkDeleting_InvalidID(t *testing.T) {
	s := setup()
	ctx := context.Background()

	err := s.service.MarkDeleting(ctx, 0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user id")
}

func TestFinalizeDelete_Success(t *testing.T) {
	s := setup()
	ctx := context.Background()

	mockTx := new(MockTx)
	s.mockPool.On("Begin", ctx).Return(mockTx, nil)
	mockTx.On("Rollback", ctx).Return(nil)
	mockTx.On("Commit", ctx).Return(nil)
	s.mockRepo.On("GetUser", ctx, 1).Return(domain.User{
		ID:     1,
		Email:  "john@example.com",
		Status: "deleting",
	}, nil)
	s.mockRepo.On("DeleteUserTx", ctx, mockTx, 1).Return(nil)
	s.mockPublisher.On("Publish", ctx, "user.deleted", mock.Anything).Return(nil)

	err := s.service.FinalizeDelete(ctx, 1)

	assert.NoError(t, err)
	s.mockRepo.AssertExpectations(t)
}

func TestRestoreUser_Success(t *testing.T) {
	s := setup()
	ctx := context.Background()

	mockTx := new(MockTx)
	s.mockPool.On("Begin", ctx).Return(mockTx, nil)
	mockTx.On("Rollback", ctx).Return(nil)
	mockTx.On("Commit", ctx).Return(nil)
	s.mockRepo.On("UpdateStatusTx", ctx, mockTx, 1, "active").Return(nil)

	err := s.service.RestoreUser(ctx, 1)

	assert.NoError(t, err)
	s.mockRepo.AssertExpectations(t)
}

func TestUpdateRole_Success(t *testing.T) {
	s := setup()
	ctx := context.Background()

	s.mockUserCache.On("GetUser", ctx, 1).Return(domain.User{
		ID:           1,
		FullName:     "John Doe",
		Email:        "john@example.com",
		PasswordHash: "hashed_password",
		IsAdmin:      false,
	}, nil)
	s.mockRepo.On("UpdateRole", ctx, 1, true).Return(domain.User{
		ID:           1,
		FullName:     "John Doe",
		Email:        "john@example.com",
		PasswordHash: "hashed_password",
		IsAdmin:      true,
	}, nil)
	s.mockUserCache.On("InvalidateUser", ctx, 1).Return(nil)
	s.mockListCache.On("InvalidateAllUsersList", ctx).Return(nil)

	user, err := s.service.UpdateRole(ctx, 1, true)

	assert.NoError(t, err)
	assert.True(t, user.IsAdmin)
	s.mockRepo.AssertExpectations(t)
}

func TestAdminExists(t *testing.T) {
	s := setup()
	ctx := context.Background()

	s.mockRedis.On("Get", ctx, "admin:exists", mock.Anything).Return(errors.New("not found"))
	s.mockRepo.On("AdminExists", ctx).Return(true, nil)
	s.mockRedis.On("Set", ctx, "admin:exists", true, 5*time.Minute).Return(nil)

	exists, err := s.service.AdminExists(ctx)

	assert.NoError(t, err)
	assert.True(t, exists)
	s.mockRepo.AssertExpectations(t)
}

func TestGetMetrics(t *testing.T) {
	s := setup()
	ctx := context.Background()

	s.mockRepo.On("GetTotalUsers", ctx).Return(100, nil)

	total, err := s.service.GetMetrics(ctx)

	assert.NoError(t, err)
	assert.Equal(t, 100, total)
	s.mockRepo.AssertExpectations(t)
}

func TestDeleteUser_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		id          int
		expectError bool
	}{
		{
			name:        "successful delete",
			id:          1,
			expectError: false,
		},
		{
			name:        "invalid id",
			id:          0,
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

			if tt.id > 0 {
				mockTx.On("Commit", ctx).Return(nil)
				s.mockRepo.On("DeleteUserTx", ctx, mockTx, tt.id).Return(nil)
				s.mockUserCache.On("InvalidateUser", ctx, tt.id).Return(nil)
				s.mockListCache.On("InvalidateAllUsersList", ctx).Return(nil)
			} else {
				s.mockRepo.On("DeleteUserTx", ctx, mockTx, tt.id).Return(errors.New("user not found"))
			}

			err := s.service.DeleteUser(ctx, tt.id)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetUser_TableDriven(t *testing.T) {
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
		{
			name:        "not found",
			userID:      999,
			cached:      false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := setup()
			ctx := context.Background()

			if tt.cached {
				s.mockUserCache.On("GetUser", ctx, tt.userID).Return(domain.User{ID: tt.userID}, nil)
			} else {
				s.mockUserCache.On("GetUser", ctx, tt.userID).Return(domain.User{}, errors.New("not found"))
				if tt.expectError {
					s.mockRepo.On("GetUser", ctx, tt.userID).Return(domain.User{}, errors.New("not found"))
				} else {
					s.mockRepo.On("GetUser", ctx, tt.userID).Return(domain.User{ID: tt.userID}, nil)
					s.mockUserCache.On("SetUser", ctx, mock.Anything).Return(nil)
				}
			}

			_, err := s.service.GetUser(ctx, tt.userID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
