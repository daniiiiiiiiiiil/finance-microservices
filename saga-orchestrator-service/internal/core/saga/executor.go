package saga

import (
	"context"
	"fmt"

	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/ports"
	"go.uber.org/zap"
)

func (m *SagaManager) executeSteps(ctx context.Context, saga *domain.Saga) error {
	for i, step := range saga.Steps {
		saga.SetCurrentStep(i)

		m.logger.Debug("executing saga step",
			zap.String("saga_id", saga.SagaID),
			zap.String("step_name", step.Name),
			zap.Int("step_order", step.Order))

		step.MarkStarted()
		if err := m.repo.UpdateStep(ctx, step); err != nil {
			m.logger.Error("failed to update step", zap.Error(err))
		}

		if err := m.executeStep(ctx, saga, step); err != nil {
			step.MarkFailed(err.Error())
			if updErr := m.repo.UpdateStep(ctx, step); updErr != nil {
				m.logger.Error("failed to update step", zap.Error(updErr))
			}

			saga.MarkFailed(err.Error())
			if updErr := m.repo.Update(ctx, saga); updErr != nil {
				m.logger.Error("failed to update step", zap.Error(updErr))
			}

			return fmt.Errorf("step %s failed: %w", step.Name, err)
		}

		step.MarkCompleted()
		if err := m.repo.UpdateStep(ctx, step); err != nil {
			m.logger.Error("failed to update step", zap.Error(err))
		}
		m.logger.Debug("saga step completed",
			zap.String("saga_id", saga.SagaID),
			zap.String("step_name", step.Name))
	}
	return nil
}

func (m *SagaManager) executeStep(ctx context.Context, saga *domain.Saga, step *domain.Step) error {
	definition, err := m.registry.Get(saga.Type)
	if err != nil {
		return fmt.Errorf("get saga definition: %w", err)
	}

	var stepDef *ports.SagaStepDefinition
	for _, s := range definition.Steps {
		if s.Name == step.Name {
			stepDefCopy := s
			stepDef = &stepDefCopy
			break
		}
	}

	if stepDef == nil {
		return fmt.Errorf("step definition not found: %s", step.Name)
	}

	return stepDef.Action(ctx, saga)
}
