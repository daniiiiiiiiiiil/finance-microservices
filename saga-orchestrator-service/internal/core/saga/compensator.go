package saga

import (
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/ports"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

func (m *SagaManager) compensate(ctx context.Context, saga *domain.Saga) error {
	saga.MarkCompensating()
	if err := m.repo.UpdateStatus(ctx, saga.ID, domain.StatusCompensating, saga.Error); err != nil {
		m.logger.Error("failed to update saga status", zap.Error(err))
	}

	m.logger.Warn("starting saga compensation",
		zap.String("saga_id", saga.SagaID),
		zap.Int("current_step", saga.CurrentStep))

	definition, err := m.registry.Get(saga.Type)
	if err != nil {
		return fmt.Errorf("saga type %q not found", saga.Type)
	}

	for i := saga.CurrentStep; i >= 0; i-- {
		step := saga.Steps[i]

		if !step.IsCompleted() {
			continue
		}

		var stepDef *ports.SagaStepDefinition
		for _, s := range definition.Steps {
			if s.Name == step.Name {
				stepDefCopy := s
				stepDef = &stepDefCopy
				break
			}
		}
		if stepDef == nil || stepDef.Compensate == nil {
			continue
		}

		comp := domain.NewCompensation(saga.ID, step.Name, step.Metadata)
		comp.MarkInProgress()
		if err := m.repo.SaveCompensation(ctx, comp); err != nil {
			m.logger.Error("failed to save compensation", zap.Error(err))
		}

		if err := stepDef.Compensate(ctx, saga); err != nil {
			m.logger.Error("saga step compensation failed",
				zap.String("saga_id", saga.SagaID),
				zap.String("step_name", step.Name),
				zap.Error(err))

			comp.MarkFailed(err.Error())
			if updErr := m.repo.UpdateCompensation(ctx, comp); updErr != nil {
				m.logger.Error("failed to update compensation", zap.Error(updErr))
			}
			continue
		}

		comp.MarkCompleted()
		if err := m.repo.UpdateCompensation(ctx, comp); err != nil {
			m.logger.Error("failed to update compensation", zap.Error(err))
		}

		step.MarkCompensated()
		if err := m.repo.UpdateStep(ctx, step); err != nil {
			m.logger.Error("failed to update step", zap.Error(err))
		}
	}
	saga.MarkCompensated()
	if err := m.repo.UpdateStatus(ctx, saga.ID, domain.StatusCompensated, saga.Error); err != nil {
		m.logger.Error("failed to update saga status", zap.Error(err))
	}
	m.removeSaga(saga.SagaID)

	return nil
}
