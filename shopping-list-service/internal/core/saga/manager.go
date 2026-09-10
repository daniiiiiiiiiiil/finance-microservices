package saga

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/pkg/logger"
	"go.uber.org/zap"
)

type Saga struct {
	ID          string
	Type        string
	UserID      int
	Status      Status
	Steps       []*Step
	CurrentStep int
	Error       error
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt *time.Time
	Metadata    map[string]interface{}
}

func NewSaga(id, sagaType string, userId int, steps []*Step) *Saga {
	return &Saga{
		ID:          id,
		Type:        sagaType,
		UserID:      userId,
		Status:      StatusPending,
		Steps:       steps,
		CurrentStep: 0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    make(map[string]interface{}),
	}
}

type SagaManager struct {
	mu     sync.RWMutex
	sagas  map[string]*Saga
	logger *logger.Logger
	repo   SagaRepository
}

func NewSagaManager(logger *logger.Logger, repo SagaRepository) *SagaManager {
	return &SagaManager{
		sagas:  make(map[string]*Saga),
		logger: logger,
		repo:   repo,
	}
}

type SagaRepository interface {
	Save(ctx context.Context, saga *Saga) error
	Load(ctx context.Context, sagaID string) (*Saga, error)
	UpdateStatus(ctx context.Context, sagaID string, status Status, err error) error
	GetActiveSagas(ctx context.Context) ([]*Saga, error)
}

func (m *SagaManager) StartSaga(ctx context.Context, saga *Saga) error {
	m.mu.Lock()
	m.sagas[saga.ID] = saga
	m.mu.Unlock()

	m.logger.Info("Saga started",
		zap.String("saga_id", saga.ID),
		zap.String("type", saga.Type),
		zap.Int("user_id", saga.UserID),
		zap.Int("total_steps", len(saga.Steps)))

	if m.repo != nil {
		if err := m.repo.Save(ctx, saga); err != nil {
			m.logger.Error("failed to save saga to repository",
				zap.String("saga_id", saga.ID),
				zap.Error(err))
		}
	}
	saga.Status = StatusInProgress
	saga.UpdatedAt = time.Now()

	if err := m.executeSteps(ctx, saga); err != nil {
		m.logger.Error("failed to execute steps",
			zap.String("saga_id", saga.ID),
			zap.Error(err))
		if compErr := m.compensate(ctx, saga); compErr != nil {
			m.logger.Error("saga compensation failed",
				zap.String("saga_id", saga.ID),
				zap.Error(compErr))
		}

		return err
	}
	now := time.Now()
	saga.Status = StatusCompleted
	saga.CompletedAt = &now
	saga.UpdatedAt = now

	m.logger.Info("saga completed successfully",
		zap.String("saga_id", saga.ID),
		zap.Duration("duration", now.Sub(saga.CreatedAt)))

	if m.repo != nil {
		if err := m.repo.UpdateStatus(ctx, saga.ID, saga.Status, nil); err != nil {
			m.logger.Error("failed to update saga status in repository",
				zap.String("saga_id", saga.ID),
				zap.Error(err))
		}
	}
	m.removeSaga(saga.ID)
	return nil
}

func (m *SagaManager) executeSteps(ctx context.Context, saga *Saga) error {
	for i, step := range saga.Steps {
		saga.CurrentStep = i

		m.logger.Debug("executing saga step",
			zap.String("saga_id", saga.ID),
			zap.String("step_name", step.Name),
			zap.Int("step_order", step.Order))

		step.MarkStarted()
		saga.UpdatedAt = time.Now()

		if err := step.Action(ctx); err != nil {
			step.MarkFailed(err)
			saga.Status = StatusFailed
			saga.Error = err
			saga.UpdatedAt = time.Now()

			m.logger.Error("saga step failed",
				zap.String("saga_id", saga.ID),
				zap.String("step_name", step.Name),
				zap.Error(err))

			if m.repo != nil {
				if err := m.repo.UpdateStatus(ctx, saga.ID, saga.Status, err); err != nil {
					m.logger.Error("failed to update saga status",
						zap.String("saga_id", saga.ID),
						zap.Error(err))
				}
			}
			return fmt.Errorf("step %s failed: %w", step.Name, err)
		}

		step.MarkCompleted()
		saga.UpdatedAt = time.Now()

		m.logger.Debug("saga step completed",
			zap.String("saga_id", saga.ID),
			zap.String("step_name", step.Name))
	}
	return nil
}

func (m *SagaManager) compensate(ctx context.Context, saga *Saga) error {
	saga.Status = StatusCompensating
	saga.UpdatedAt = time.Now()

	m.logger.Warn("starting saga compensation",
		zap.String("saga_id", saga.ID),
		zap.Int("current_step", saga.CurrentStep))

	for i := saga.CurrentStep; i >= 0; i-- {
		step := saga.Steps[i]
		if !step.CanCompensate() {
			continue
		}
		m.logger.Debug("compensating saga step",
			zap.String("saga_id", saga.ID),
			zap.String("step_name", step.Name))
		if err := step.Compensate(ctx); err != nil {
			m.logger.Error("saga step compensation failed",
				zap.String("saga_id", saga.ID),
				zap.String("step_name", step.Name),
				zap.Error(err))
		} else {
			step.MarkCompensated()
			m.logger.Debug("saga step compensated",
				zap.String("saga_id", saga.ID),
				zap.String("step_name", step.Name))
		}
	}
	saga.Status = StatusCompensated
	now := time.Now()
	saga.CompletedAt = &now
	saga.UpdatedAt = now

	m.logger.Info("saga compensation completed",
		zap.String("saga_id", saga.ID))
	if m.repo != nil {
		if err := m.repo.UpdateStatus(ctx, saga.ID, saga.Status, nil); err != nil {
			m.logger.Error("failed to update saga status",
				zap.String("saga_id", saga.ID),
				zap.Error(err))
		}
	}
	m.removeSaga(saga.ID)
	return nil
}

func (m *SagaManager) GetSaga(sagaID string) (*Saga, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	saga, ok := m.sagas[sagaID]
	return saga, ok
}

func (m *SagaManager) GetActiveSagas() []*Saga {
	m.mu.RLock()
	defer m.mu.RUnlock()
	sagas := make([]*Saga, 0, len(m.sagas))
	for _, saga := range m.sagas {
		if !saga.Status.IsFinal() {
			sagas = append(sagas, saga)
		}
	}
	return sagas
}

func (m *SagaManager) removeSaga(sagaID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sagas, sagaID)
}

func (m *SagaManager) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Info("shutting down saga manager",
		zap.Int("active_sagas", len(m.sagas)))

	for _, saga := range m.sagas {
		m.logger.Warn("active saga on shutdown",
			zap.String("saga_id", saga.ID),
			zap.String("status", string(saga.Status)),
			zap.Int("current_step", saga.CurrentStep))
	}
	return nil
}
