package saga

import (
	"context"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/domain"
	"go.uber.org/zap"
)

func (m *SagaManager) LoadState(ctx context.Context, sagaID string) (*domain.Saga, error) {
	saga, err := m.repo.GetBySagaID(ctx, sagaID)
	if err != nil {
		return nil, fmt.Errorf("cannot load state for saga %s: %w", sagaID, err)
	}
	return saga, nil
}

func (m *SagaManager) SaveState(ctx context.Context, state *domain.Saga) error {
	if err := m.repo.Update(ctx, state); err != nil {
		return fmt.Errorf("cannot save state for saga %s: %w", state.SagaID, err)
	}
	return nil
}

func (m *SagaManager) ResumeActiveSagas(ctx context.Context) error {
	sagas, err := m.repo.GetActive(ctx)
	if err != nil {
		return fmt.Errorf("cannot load active sagas: %w", err)
	}

	m.logger.Info("resuming active sagas", zap.Int("count", len(sagas)))

	for _, saga := range sagas {
		m.logger.Info("resuming saga",
			zap.String("saga_id", saga.SagaID),
			zap.String("status", saga.Status.String()),
			zap.Int("current_step", saga.CurrentStep))

		go func(s *domain.Saga) {
			if err := m.executeSteps(ctx, s); err != nil {
				m.logger.Warn("cannot execute steps for saga", zap.String("saga_id", s.SagaID), zap.Error(err))

				if compErr := m.compensate(ctx, s); compErr != nil {
					m.logger.Error("failed to compensate saga",
						zap.String("saga_id", s.SagaID),
						zap.Error(compErr))
				}
				return
			}
			s.MarkCompleted()

			if err := m.repo.UpdateStatus(ctx, s.ID, domain.StatusCompleted, ""); err != nil {
				m.logger.Warn("cannot update status for saga")
			}
			if err := m.publisher.SendUserDeleteCompleted(ctx, s.UserID, s.SagaID); err != nil {
				m.logger.Warn("cannot update status for saga")
			}
			m.removeSaga(s.SagaID)
		}(saga)
	}
	return nil
}
