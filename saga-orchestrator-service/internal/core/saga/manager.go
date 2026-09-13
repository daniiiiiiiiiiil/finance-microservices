package saga

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/pkg/logger"
	"go.uber.org/zap"
)

type SagaManager struct {
	mtx       sync.RWMutex
	sagas     map[string]*domain.Saga
	logger    *logger.Logger
	repo      ports.SagaRepository
	registry  ports.SagaRegistry
	publisher ports.EventPublisher
}

func NewSagaManager(
	logger *logger.Logger,
	repo ports.SagaRepository,
	registry ports.SagaRegistry,
	publisher ports.EventPublisher,
) *SagaManager {
	return &SagaManager{
		sagas:     make(map[string]*domain.Saga),
		logger:    logger,
		repo:      repo,
		registry:  registry,
		publisher: publisher,
	}
}

func (m *SagaManager) StartSaga(ctx context.Context, saga *domain.Saga) error {
	m.logger.Info("starting saga",
		zap.String("saga_id", saga.SagaID),
		zap.String("type", saga.Type.String()),
		zap.Int("user_id", saga.UserID),
		zap.Int("total_steps", saga.TotalSteps))

	if err := saga.Validate(); err != nil {
		return fmt.Errorf("validate saga: %w", err)
	}

	existing, err := m.repo.GetBySagaID(ctx, saga.SagaID)
	if err == nil && existing != nil {
		m.logger.Warn("saga already exists, skipping",
			zap.String("saga_id", saga.SagaID))
		return nil
	}

	if err := m.repo.Save(ctx, saga); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			m.logger.Warn("saga already saved, skipping",
				zap.String("saga_id", saga.SagaID))
			return nil
		}
		return fmt.Errorf("save saga: %w", err)
	}

	m.mtx.Lock()
	m.sagas[saga.SagaID] = saga
	m.mtx.Unlock()

	definition, err := m.registry.Get(saga.Type)
	if err != nil {
		return fmt.Errorf("get definition of saga: %w", err)
	}

	if err := definition.Build(ctx, saga); err != nil {
		return fmt.Errorf("build saga: %w", err)
	}

	saga.MarkInProgress()
	if err := m.repo.UpdateStatus(ctx, saga.ID, domain.StatusInProgress, ""); err != nil {
		m.logger.Error("failed to update status", zap.Error(err))
	}

	if err := m.executeSteps(ctx, saga); err != nil {
		m.logger.Error("saga execution failed",
			zap.String("saga_id", saga.SagaID),
			zap.Error(err))

		if compErr := m.compensate(ctx, saga); compErr != nil {
			m.logger.Error("saga compensation failed",
				zap.String("saga_id", saga.SagaID),
				zap.Error(compErr))
		}

		if pubErr := m.publisher.SendUserDeleteFailed(ctx, saga.UserID, saga.SagaID, "saga execution failed", err.Error()); pubErr != nil {
			m.logger.Error("failed to publish failed event", zap.Error(pubErr))
		}

		return err
	}
	saga.MarkCompleted()
	if err := m.repo.UpdateStatus(ctx, saga.ID, domain.StatusCompleted, ""); err != nil {
		m.logger.Error("failed to update status", zap.Error(err))
	}

	if err := m.publisher.SendUserDeleteCompleted(ctx, saga.UserID, saga.SagaID); err != nil {
		m.logger.Error("failed to publish completed event", zap.Error(err))
	}

	m.logger.Info("saga completed successfully",
		zap.String("saga_id", saga.SagaID),
		zap.Duration("duration", time.Since(saga.CreatedAt)))

	m.removeSaga(saga.SagaID)
	return nil
}

func (m *SagaManager) GetSaga(sagaID string) (*domain.Saga, bool) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	saga, ok := m.sagas[sagaID]
	return saga, ok
}

func (m *SagaManager) GetActiveSagas() []*domain.Saga {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	sagas := make([]*domain.Saga, 0, len(m.sagas))
	for _, saga := range m.sagas {
		if saga.IsActive() {
			sagas = append(sagas, saga)
		}
	}
	return sagas
}

func (m *SagaManager) removeSaga(sagaID string) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	delete(m.sagas, sagaID)
}

func (m *SagaManager) Shutdown(ctx context.Context) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	m.logger.Info("shutting down saga manager",
		zap.Int("active_sagas", len(m.sagas)))

	for _, saga := range m.sagas {
		m.logger.Warn("active saga on shutdown",
			zap.String("saga_id", saga.SagaID),
			zap.String("status", saga.Status.String()),
			zap.Int("current_step", saga.CurrentStep))
	}
	return nil
}
