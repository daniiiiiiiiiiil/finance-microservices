package orchestrator

import (
	"context"
	"errors"
	"testing"

	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/repository/postgres/pool"
	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockSagaManager struct {
	mock.Mock
}

func (m *MockSagaManager) StartSaga(ctx context.Context, saga *domain.Saga) error {
	return m.Called(ctx, saga).Error(0)
}

func (m *MockSagaManager) GetSaga(sagaID string) (*domain.Saga, bool) {
	args := m.Called(sagaID)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(*domain.Saga), args.Bool(1)
}

func (m *MockSagaManager) GetActiveSagas() []*domain.Saga {
	args := m.Called()
	return args.Get(0).([]*domain.Saga)
}

func (m *MockSagaManager) Shutdown(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) Publish(ctx context.Context, eventType string, data interface{}) error {
	return m.Called(ctx, eventType, data).Error(0)
}

func (m *MockEventPublisher) SendUserMarkDeleting(ctx context.Context, userID int, sagaID string) error {
	return m.Called(ctx, userID, sagaID).Error(0)
}

func (m *MockEventPublisher) SendUserRestore(ctx context.Context, userID int, sagaID string) error {
	return m.Called(ctx, userID, sagaID).Error(0)
}

func (m *MockEventPublisher) SendUserFinalizeDelete(ctx context.Context, userID int, sagaID string) error {
	return m.Called(ctx, userID, sagaID).Error(0)
}

func (m *MockEventPublisher) SendUserDeleteCompleted(ctx context.Context, userID int, sagaID string) error {
	return m.Called(ctx, userID, sagaID).Error(0)
}

func (m *MockEventPublisher) SendUserDeleteFailed(ctx context.Context, userID int, sagaID string, reason string, errMsg string) error {
	return m.Called(ctx, userID, sagaID, reason, errMsg).Error(0)
}

func (m *MockEventPublisher) SendDeleteShoppingData(ctx context.Context, userID int, sagaID string) error {
	return m.Called(ctx, userID, sagaID).Error(0)
}

func (m *MockEventPublisher) SendDeleteFinanceTransactions(ctx context.Context, userID int, sagaID string) error {
	return m.Called(ctx, userID, sagaID).Error(0)
}

type MockSagaRepository struct {
	mock.Mock
}

func (m *MockSagaRepository) Save(ctx context.Context, saga *domain.Saga) error {
	return m.Called(ctx, saga).Error(0)
}

func (m *MockSagaRepository) SaveTx(ctx context.Context, tx pool.Tx, saga *domain.Saga) error {
	return m.Called(ctx, tx, saga).Error(0)
}

func (m *MockSagaRepository) Update(ctx context.Context, saga *domain.Saga) error {
	return m.Called(ctx, saga).Error(0)
}

func (m *MockSagaRepository) UpdateStatus(ctx context.Context, sagaID int, status domain.Status, errMsg string) error {
	return m.Called(ctx, sagaID, status, errMsg).Error(0)
}

func (m *MockSagaRepository) UpdateStep(ctx context.Context, step *domain.Step) error {
	return m.Called(ctx, step).Error(0)
}

func (m *MockSagaRepository) SaveCompensation(ctx context.Context, comp *domain.Compensation) error {
	return m.Called(ctx, comp).Error(0)
}

func (m *MockSagaRepository) UpdateCompensation(ctx context.Context, comp *domain.Compensation) error {
	return m.Called(ctx, comp).Error(0)
}

func (m *MockSagaRepository) GetByID(ctx context.Context, id int) (*domain.Saga, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Saga), args.Error(1)
}

func (m *MockSagaRepository) GetBySagaID(ctx context.Context, sagaID string) (*domain.Saga, error) {
	args := m.Called(ctx, sagaID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Saga), args.Error(1)
}

func (m *MockSagaRepository) GetByUserID(ctx context.Context, userID int) ([]*domain.Saga, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Saga), args.Error(1)
}

func (m *MockSagaRepository) GetActive(ctx context.Context) ([]*domain.Saga, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Saga), args.Error(1)
}

func (m *MockSagaRepository) GetByStatus(ctx context.Context, status domain.Status) ([]*domain.Saga, error) {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Saga), args.Error(1)
}

func (m *MockSagaRepository) Delete(ctx context.Context, id int) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockSagaRepository) SaveOutboxTx(ctx context.Context, tx pool.Tx, event domain.OutboxEvent) error {
	return m.Called(ctx, tx, event).Error(0)
}

func (m *MockSagaRepository) SaveOutbox(ctx context.Context, event domain.OutboxEvent) error {
	return m.Called(ctx, event).Error(0)
}

func (m *MockSagaRepository) GetPendingOutbox(ctx context.Context, limit int) ([]domain.OutboxEvent, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.OutboxEvent), args.Error(1)
}

func (m *MockSagaRepository) MarkOutboxProcessed(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockSagaRepository) MarkOutboxFailed(ctx context.Context, id string, errMsg string) error {
	return m.Called(ctx, id, errMsg).Error(0)
}

func newTestLogger() *logger.Logger {
	return &logger.Logger{Logger: zap.NewNop()}
}

func newTestSaga() *domain.Saga {
	s := domain.NewSaga("test-saga-1", domain.SagaTypeDeleteUser, 42, 4)
	s.ID = 1
	return s
}

func TestSagaRegistry_Register_Success(t *testing.T) {
	reg := NewSagaRegistryImpl()

	def := ports.SagaDefinition{
		Type:  domain.SagaTypeDeleteUser,
		Build: func(ctx context.Context, saga *domain.Saga) error { return nil },
		Steps: []ports.SagaStepDefinition{
			{Name: "step1", Order: 1, Action: func(ctx context.Context, saga *domain.Saga) error { return nil }},
		},
	}

	err := reg.Register(domain.SagaTypeDeleteUser, def)
	assert.NoError(t, err)
	assert.Equal(t, 1, reg.Count())
	assert.True(t, reg.IsRegistered(domain.SagaTypeDeleteUser))
}

func TestSagaRegistry_Register_InvalidType(t *testing.T) {
	reg := NewSagaRegistryImpl()

	def := ports.SagaDefinition{
		Build: func(ctx context.Context, saga *domain.Saga) error { return nil },
		Steps: []ports.SagaStepDefinition{{Name: "s1", Action: func(ctx context.Context, saga *domain.Saga) error { return nil }}},
	}

	err := reg.Register("invalid_type", def)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid saga type")
}

func TestSagaRegistry_Register_Duplicate(t *testing.T) {
	reg := NewSagaRegistryImpl()

	def := ports.SagaDefinition{
		Type:  domain.SagaTypeDeleteUser,
		Build: func(ctx context.Context, saga *domain.Saga) error { return nil },
		Steps: []ports.SagaStepDefinition{{Name: "s1", Action: func(ctx context.Context, saga *domain.Saga) error { return nil }}},
	}

	assert.NoError(t, reg.Register(domain.SagaTypeDeleteUser, def))

	err := reg.Register(domain.SagaTypeDeleteUser, def)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestSagaRegistry_Register_MissingBuild(t *testing.T) {
	reg := NewSagaRegistryImpl()

	def := ports.SagaDefinition{
		Type:  domain.SagaTypeDeleteUser,
		Steps: []ports.SagaStepDefinition{{Name: "s1", Action: func(ctx context.Context, saga *domain.Saga) error { return nil }}},
	}

	err := reg.Register(domain.SagaTypeDeleteUser, def)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing a build")
}

func TestSagaRegistry_Register_NoSteps(t *testing.T) {
	reg := NewSagaRegistryImpl()

	def := ports.SagaDefinition{
		Type:  domain.SagaTypeDeleteUser,
		Build: func(ctx context.Context, saga *domain.Saga) error { return nil },
	}

	err := reg.Register(domain.SagaTypeDeleteUser, def)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no steps")
}

func TestSagaRegistry_Register_EmptyStepName(t *testing.T) {
	reg := NewSagaRegistryImpl()

	def := ports.SagaDefinition{
		Type:  domain.SagaTypeDeleteUser,
		Build: func(ctx context.Context, saga *domain.Saga) error { return nil },
		Steps: []ports.SagaStepDefinition{{Name: "", Order: 1, Action: func(ctx context.Context, saga *domain.Saga) error { return nil }}},
	}

	err := reg.Register(domain.SagaTypeDeleteUser, def)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no step name")
}

func TestSagaRegistry_Register_DuplicateStepName(t *testing.T) {
	reg := NewSagaRegistryImpl()

	def := ports.SagaDefinition{
		Type:  domain.SagaTypeDeleteUser,
		Build: func(ctx context.Context, saga *domain.Saga) error { return nil },
		Steps: []ports.SagaStepDefinition{
			{Name: "s1", Order: 1, Action: func(ctx context.Context, saga *domain.Saga) error { return nil }},
			{Name: "s1", Order: 2, Action: func(ctx context.Context, saga *domain.Saga) error { return nil }},
		},
	}

	err := reg.Register(domain.SagaTypeDeleteUser, def)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate step name")
}

func TestSagaRegistry_Register_NoAction(t *testing.T) {
	reg := NewSagaRegistryImpl()

	def := ports.SagaDefinition{
		Type:  domain.SagaTypeDeleteUser,
		Build: func(ctx context.Context, saga *domain.Saga) error { return nil },
		Steps: []ports.SagaStepDefinition{{Name: "s1", Order: 1, Action: nil}},
	}

	err := reg.Register(domain.SagaTypeDeleteUser, def)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no action")
}

func TestSagaRegistry_Get_Success(t *testing.T) {
	reg := NewSagaRegistryImpl()

	def := ports.SagaDefinition{
		Type:  domain.SagaTypeDeleteUser,
		Build: func(ctx context.Context, saga *domain.Saga) error { return nil },
		Steps: []ports.SagaStepDefinition{{Name: "s1", Action: func(ctx context.Context, saga *domain.Saga) error { return nil }}},
	}
	_ = reg.Register(domain.SagaTypeDeleteUser, def)

	got, err := reg.Get(domain.SagaTypeDeleteUser)
	assert.NoError(t, err)
	assert.Equal(t, def.Type, got.Type)
}

func TestSagaRegistry_Get_NotFound(t *testing.T) {
	reg := NewSagaRegistryImpl()
	_, err := reg.Get(domain.SagaTypeDeleteUser)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not registered")
}

func TestDeleteUserSaga_Definition(t *testing.T) {
	s := NewDeleteUserSaga(
		newTestLogger(),
		new(MockSagaManager),
		new(MockEventPublisher),
		new(MockSagaRepository),
	)

	def := s.Definition()

	assert.Equal(t, "delete_user", string(def.Type))
	assert.Len(t, def.Steps, 4)
	assert.Equal(t, "mark_deleting", def.Steps[0].Name)
	assert.Equal(t, "delete_shopping", def.Steps[1].Name)
	assert.Equal(t, "delete_transactions", def.Steps[2].Name)
	assert.Equal(t, "finalize_delete", def.Steps[3].Name)
}

func TestDeleteUserSaga_HandleUserDeleted_Success(t *testing.T) {
	mockMgr := new(MockSagaManager)
	s := NewDeleteUserSaga(
		newTestLogger(),
		mockMgr,
		new(MockEventPublisher),
		new(MockSagaRepository),
	)

	ctx := context.Background()
	mockMgr.On("StartSaga", ctx, mock.Anything).Return(nil)

	err := s.HandleUserDeleted(ctx, 42)

	assert.NoError(t, err)
	mockMgr.AssertCalled(t, "StartSaga", ctx, mock.Anything)
}

func TestDeleteUserSaga_HandleUserDeleted_StartError(t *testing.T) {
	mockMgr := new(MockSagaManager)
	s := NewDeleteUserSaga(
		newTestLogger(),
		mockMgr,
		new(MockEventPublisher),
		new(MockSagaRepository),
	)

	ctx := context.Background()
	mockMgr.On("StartSaga", ctx, mock.Anything).Return(errors.New("some error"))

	err := s.HandleUserDeleted(ctx, 42)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "start saga")
}

func TestDeleteUserSaga_MarkDeleting_Success(t *testing.T) {
	mockPub := new(MockEventPublisher)
	s := NewDeleteUserSaga(
		newTestLogger(),
		new(MockSagaManager),
		mockPub,
		new(MockSagaRepository),
	)

	ctx := context.Background()
	saga := newTestSaga()

	mockPub.On("SendUserMarkDeleting", ctx, saga.UserID, saga.SagaID).Return(nil)

	err := s.markDeleting(ctx, saga)
	assert.NoError(t, err)
	mockPub.AssertExpectations(t)
}

func TestDeleteUserSaga_MarkDeleting_Error(t *testing.T) {
	mockPub := new(MockEventPublisher)
	s := NewDeleteUserSaga(
		newTestLogger(),
		new(MockSagaManager),
		mockPub,
		new(MockSagaRepository),
	)

	ctx := context.Background()
	saga := newTestSaga()

	mockPub.On("SendUserMarkDeleting", ctx, saga.UserID, saga.SagaID).Return(errors.New("kafka down"))

	err := s.markDeleting(ctx, saga)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mark deleting")
}

func TestDeleteUserSaga_CompensateMarkDeleting_Success(t *testing.T) {
	mockPub := new(MockEventPublisher)
	s := NewDeleteUserSaga(
		newTestLogger(),
		new(MockSagaManager),
		mockPub,
		new(MockSagaRepository),
	)

	ctx := context.Background()
	saga := newTestSaga()

	mockPub.On("SendUserRestore", ctx, saga.UserID, saga.SagaID).Return(nil)

	err := s.compensateMarkDeleting(ctx, saga)
	assert.NoError(t, err)
	mockPub.AssertExpectations(t)
}

func TestDeleteUserSaga_DeleteShopping_Success(t *testing.T) {
	mockPub := new(MockEventPublisher)
	s := NewDeleteUserSaga(
		newTestLogger(),
		new(MockSagaManager),
		mockPub,
		new(MockSagaRepository),
	)

	ctx := context.Background()
	saga := newTestSaga()

	mockPub.On("SendDeleteShoppingData", ctx, saga.UserID, saga.SagaID).Return(nil)

	err := s.deleteShopping(ctx, saga)
	assert.NoError(t, err)
	mockPub.AssertExpectations(t)
}

func TestDeleteUserSaga_DeleteTransactions_Success(t *testing.T) {
	mockPub := new(MockEventPublisher)
	s := NewDeleteUserSaga(
		newTestLogger(),
		new(MockSagaManager),
		mockPub,
		new(MockSagaRepository),
	)

	ctx := context.Background()
	saga := newTestSaga()

	mockPub.On("SendDeleteFinanceTransactions", ctx, saga.UserID, saga.SagaID).Return(nil)

	err := s.deleteTransactions(ctx, saga)
	assert.NoError(t, err)
	mockPub.AssertExpectations(t)
}

func TestDeleteUserSaga_FinalizeDelete_Success(t *testing.T) {
	mockPub := new(MockEventPublisher)
	s := NewDeleteUserSaga(
		newTestLogger(),
		new(MockSagaManager),
		mockPub,
		new(MockSagaRepository),
	)

	ctx := context.Background()
	saga := newTestSaga()

	mockPub.On("SendUserFinalizeDelete", ctx, saga.UserID, saga.SagaID).Return(nil)

	err := s.finalizeDelete(ctx, saga)
	assert.NoError(t, err)
	mockPub.AssertExpectations(t)
}

func TestRegisterUserSaga_Definition(t *testing.T) {
	s := NewRegisterUserSaga(
		newTestLogger(),
		new(MockSagaManager),
		new(MockEventPublisher),
		new(MockSagaRepository),
	)

	def := s.Definition()

	assert.Equal(t, "register_user", string(def.Type))
	assert.Len(t, def.Steps, 3)
	assert.Equal(t, "create_credentials", def.Steps[0].Name)
	assert.Equal(t, "create_profile", def.Steps[1].Name)
	assert.Equal(t, "activate_account", def.Steps[2].Name)
}

func TestRegisterUserSaga_HandleRegisterRequested_ValidationErrors(t *testing.T) {
	s := NewRegisterUserSaga(
		newTestLogger(),
		new(MockSagaManager),
		new(MockEventPublisher),
		new(MockSagaRepository),
	)

	ctx := context.Background()

	tests := []struct {
		name     string
		email    string
		fullName string
		hash     string
	}{
		{"empty email", "", "Name", "hash"},
		{"empty name", "a@b.com", "", "hash"},
		{"empty hash", "a@b.com", "Name", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.HandleRegisterRequested(ctx, tt.email, tt.fullName, tt.hash, nil, false)
			assert.Error(t, err)
		})
	}
}

func TestRegisterUserSaga_HandleRegisterRequested_Success(t *testing.T) {
	mockMgr := new(MockSagaManager)
	s := NewRegisterUserSaga(
		newTestLogger(),
		mockMgr,
		new(MockEventPublisher),
		new(MockSagaRepository),
	)

	ctx := context.Background()
	mockMgr.On("StartSaga", ctx, mock.Anything).Return(nil)

	err := s.HandleRegisterRequested(ctx, "a@b.com", "Name", "hash", nil, false)

	assert.NoError(t, err)
	mockMgr.AssertExpectations(t)
}

func TestRegisterUserSaga_CreateCredentials(t *testing.T) {
	mockPub := new(MockEventPublisher)
	s := NewRegisterUserSaga(
		newTestLogger(),
		new(MockSagaManager),
		mockPub,
		new(MockSagaRepository),
	)

	ctx := context.Background()
	saga := domain.NewSaga("test-reg", domain.SagaTypeRegisterUser, 0, 3)
	saga.SetMetadata("email", "a@b.com")
	saga.SetMetadata("password_hash", "hash")

	mockPub.On("Publish", ctx, "auth.create_credentials", mock.Anything).Return(nil)

	err := s.createCredentials(ctx, saga)
	assert.NoError(t, err)
	mockPub.AssertExpectations(t)
}

func TestRegisterUserSaga_CreateProfile(t *testing.T) {
	mockPub := new(MockEventPublisher)
	s := NewRegisterUserSaga(
		newTestLogger(),
		new(MockSagaManager),
		mockPub,
		new(MockSagaRepository),
	)

	ctx := context.Background()
	saga := domain.NewSaga("test-reg", domain.SagaTypeRegisterUser, 0, 3)
	saga.SetMetadata("email", "a@b.com")
	saga.SetMetadata("full_name", "Name")
	saga.SetMetadata("is_admin", false)

	mockPub.On("Publish", ctx, "user.create_profile", mock.Anything).Return(nil)

	err := s.createProfile(ctx, saga)
	assert.NoError(t, err)
	mockPub.AssertExpectations(t)
}

func TestRegisterUserSaga_ActivateAccount(t *testing.T) {
	mockPub := new(MockEventPublisher)
	s := NewRegisterUserSaga(
		newTestLogger(),
		new(MockSagaManager),
		mockPub,
		new(MockSagaRepository),
	)

	ctx := context.Background()
	saga := domain.NewSaga("test-reg", domain.SagaTypeRegisterUser, 0, 3)
	saga.SetMetadata("email", "a@b.com")

	mockPub.On("Publish", ctx, "auth.activate_account", mock.Anything).Return(nil)

	err := s.activateAccount(ctx, saga)
	assert.NoError(t, err)
	mockPub.AssertExpectations(t)
}

func TestRegisterUserSaga_CompensateCreateCredentials(t *testing.T) {
	mockPub := new(MockEventPublisher)
	s := NewRegisterUserSaga(
		newTestLogger(),
		new(MockSagaManager),
		mockPub,
		new(MockSagaRepository),
	)

	ctx := context.Background()
	saga := domain.NewSaga("test-reg", domain.SagaTypeRegisterUser, 0, 3)
	saga.SetMetadata("email", "a@b.com")

	mockPub.On("Publish", ctx, "auth.delete_credentials", mock.Anything).Return(nil)

	err := s.compensateCreateCredentials(ctx, saga)
	assert.NoError(t, err)
	mockPub.AssertExpectations(t)
}

func TestRegisterUserSaga_CompensateCreateProfile(t *testing.T) {
	mockPub := new(MockEventPublisher)
	s := NewRegisterUserSaga(
		newTestLogger(),
		new(MockSagaManager),
		mockPub,
		new(MockSagaRepository),
	)

	ctx := context.Background()
	saga := domain.NewSaga("test-reg", domain.SagaTypeRegisterUser, 0, 3)
	saga.SetMetadata("email", "a@b.com")

	mockPub.On("Publish", ctx, "user.delete_profile", mock.Anything).Return(nil)

	err := s.compensateCreateProfile(ctx, saga)
	assert.NoError(t, err)
	mockPub.AssertExpectations(t)
}

func TestCompensationHandler_SaveCompensation_Success(t *testing.T) {
	mockRepo := new(MockSagaRepository)
	h := NewCompensationHandler(newTestLogger(), mockRepo, new(MockEventPublisher))

	ctx := context.Background()
	saga := newTestSaga()
	step := domain.NewStep(saga.ID, "step1", 1)

	mockRepo.On("SaveCompensation", ctx, mock.Anything).Return(nil)

	comp, err := h.SaveCompensation(ctx, saga, step, domain.CompensationStatusInProgress, "")

	assert.NoError(t, err)
	assert.NotNil(t, comp)
	mockRepo.AssertExpectations(t)
}

func TestCompensationHandler_SaveCompensation_Error(t *testing.T) {
	mockRepo := new(MockSagaRepository)
	h := NewCompensationHandler(newTestLogger(), mockRepo, new(MockEventPublisher))

	ctx := context.Background()
	saga := newTestSaga()
	step := domain.NewStep(saga.ID, "step1", 1)

	mockRepo.On("SaveCompensation", ctx, mock.Anything).Return(errors.New("db down"))

	_, err := h.SaveCompensation(ctx, saga, step, domain.CompensationStatusInProgress, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "save compensation")
}

func TestCompensationHandler_ShouldCompensate_Completed(t *testing.T) {
	h := NewCompensationHandler(newTestLogger(), new(MockSagaRepository), new(MockEventPublisher))

	step := &domain.Step{Status: domain.StepStatusCompleted}
	assert.True(t, h.ShouldCompensate(step))
}

func TestCompensationHandler_ShouldCompensate_NotCompleted(t *testing.T) {
	h := NewCompensationHandler(newTestLogger(), new(MockSagaRepository), new(MockEventPublisher))

	step := &domain.Step{Status: domain.StepStatusFailed}
	assert.False(t, h.ShouldCompensate(step))
}

func TestCompensationHandler_GetCompensatableSteps(t *testing.T) {
	h := NewCompensationHandler(newTestLogger(), new(MockSagaRepository), new(MockEventPublisher))

	saga := newTestSaga()
	saga.Steps = []*domain.Step{
		{Name: "step1", Status: domain.StepStatusCompleted},
		{Name: "step2", Status: domain.StepStatusFailed},
		{Name: "step3", Status: domain.StepStatusCompleted},
	}
	saga.CurrentStep = 2

	steps := h.GetCompensatableSteps(saga)

	assert.Len(t, steps, 2)
	assert.Equal(t, "step3", steps[0].Name)
	assert.Equal(t, "step1", steps[1].Name)
}

func TestCompensationHandler_NotifyCompensationFailure_Success(t *testing.T) {
	mockPub := new(MockEventPublisher)
	h := NewCompensationHandler(newTestLogger(), new(MockSagaRepository), mockPub)

	ctx := context.Background()
	saga := newTestSaga()
	step := domain.NewStep(saga.ID, "step1", 1)

	mockPub.On("SendUserDeleteFailed",
		ctx,
		saga.UserID,
		saga.SagaID,
		mock.Anything,
		mock.Anything,
	).Return(nil)

	err := h.NotifyCompensationFailure(ctx, saga, step, errors.New("test error"))

	assert.NoError(t, err)
	mockPub.AssertExpectations(t)
}

func TestCompensationHandler_NotifyCompensationFailure_PublishError(t *testing.T) {
	mockPub := new(MockEventPublisher)
	h := NewCompensationHandler(newTestLogger(), new(MockSagaRepository), mockPub)

	ctx := context.Background()
	saga := newTestSaga()
	step := domain.NewStep(saga.ID, "step1", 1)

	mockPub.On("SendUserDeleteFailed",
		ctx,
		saga.UserID,
		saga.SagaID,
		mock.Anything,
		mock.Anything,
	).Return(errors.New("kafka down"))

	err := h.NotifyCompensationFailure(ctx, saga, step, errors.New("test error"))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "compensation failed event")
}
