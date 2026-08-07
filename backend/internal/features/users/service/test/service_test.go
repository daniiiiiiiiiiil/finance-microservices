package test

import (
	"backend/internal/core/domain"
	"backend/internal/core/repository/postgres/pool"
	service_user "backend/internal/features/users/service"
	"backend/internal/features/users/service/mocks"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
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

func TestUsersService_GetUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUsersRepository(ctrl)
	mockPool := &MockPool{}

	service := service_user.NewUsersService(mockRepo, mockPool)

	ctx := context.Background()
	userID := 1
	expected := domain.User{ID: 1, FullName: "Test User", Email: "test@example.com"}

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

func TestUsersService_GetUser_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUsersRepository(ctrl)
	mockPool := &MockPool{}

	service := service_user.NewUsersService(mockRepo, mockPool)

	ctx := context.Background()
	userID := 999

	mockRepo.EXPECT().
		GetUser(ctx, userID).
		Return(domain.User{}, errors.New("not found")).
		Times(1)

	user, err := service.GetUser(ctx, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "userRepository.GetUser: not found")
	assert.Equal(t, domain.User{}, user)
}

func TestUsersService_PatchUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUsersRepository(ctrl)
	mockPool := &MockPool{}

	service := service_user.NewUsersService(mockRepo, mockPool)

	ctx := context.Background()
	userID := 1

	existingUser := domain.User{
		ID:           1,
		FullName:     "Old Name",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
	}
	expectedUser := domain.User{
		ID:           1,
		FullName:     "New Name",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
	}

	patch := domain.UserPatch{
		FullName: domain.Nullable[string]{Value: stringPtr("New Name"), Set: true},
	}

	mockRepo.EXPECT().
		GetUser(ctx, userID).
		Return(existingUser, nil).
		Times(1)

	mockRepo.EXPECT().
		PatchUser(ctx, userID, gomock.Any()).
		Return(expectedUser, nil).
		Times(1)

	user, err := service.PatchUser(ctx, userID, patch)

	require.NoError(t, err)
	assert.Equal(t, "New Name", user.FullName)
}

func TestUsersService_PatchUser_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUsersRepository(ctrl)
	mockPool := &MockPool{}

	service := service_user.NewUsersService(mockRepo, mockPool)

	ctx := context.Background()
	userID := 999

	patch := domain.UserPatch{
		FullName: domain.Nullable[string]{Value: stringPtr("New Name"), Set: true},
	}

	mockRepo.EXPECT().
		GetUser(ctx, userID).
		Return(domain.User{}, errors.New("not found")).
		Times(1)

	user, err := service.PatchUser(ctx, userID, patch)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "userRepository.GetUser: not found")
	assert.Equal(t, domain.User{}, user)
}

func TestUsersService_DeleteUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUsersRepository(ctrl)
	mockPool := &MockPool{}

	service := service_user.NewUsersService(mockRepo, mockPool)

	ctx := context.Background()
	userID := 1

	mockRepo.EXPECT().
		DeleteUserTx(ctx, gomock.Any(), userID).
		Return(nil).
		Times(1)

	err := service.DeleteUser(ctx, userID)

	require.NoError(t, err)
}

func TestUsersService_DeleteUser_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUsersRepository(ctrl)
	mockPool := &MockPool{}

	service := service_user.NewUsersService(mockRepo, mockPool)

	ctx := context.Background()
	userID := 999

	mockRepo.EXPECT().
		DeleteUserTx(ctx, gomock.Any(), userID).
		Return(errors.New("not found")).
		Times(1)

	err := service.DeleteUser(ctx, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DeleteUser: not found")
}

func stringPtr(s string) *string {
	return &s
}
