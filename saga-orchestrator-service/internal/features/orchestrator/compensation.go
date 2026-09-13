package orchestrator

import (
	"context"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/pkg/logger"
	"go.uber.org/zap"
)

type CompensationHandler struct {
	logger    *logger.Logger
	repo      ports.SagaRepository
	publisher ports.EventPublisher
}

func NewCompensationHandler(
	logger *logger.Logger,
	repo ports.SagaRepository,
	publisher ports.EventPublisher,
) *CompensationHandler {
	return &CompensationHandler{
		logger:    logger,
		repo:      repo,
		publisher: publisher,
	}
}

func (h *CompensationHandler) SaveCompensation(
	ctx context.Context,
	saga *domain.Saga,
	step *domain.Step,
	status domain.CompensationStatus,
	errMsg string,
) (*domain.Compensation, error) {
	comp := domain.NewCompensation(saga.ID, step.Name, step.Metadata)
	switch status {
	case domain.CompensationStatusInProgress:
		comp.MarkInProgress()
	case domain.CompensationStatusCompleted:
		comp.MarkCompleted()
	case domain.CompensationStatusFailed:
		comp.MarkFailed(errMsg)
	}
	if err := h.repo.SaveCompensation(ctx, comp); err != nil {
		return nil, fmt.Errorf("save compensation: %w", err)
	}

	h.logger.Debug("compensation saved",
		zap.String("saga_id", saga.SagaID),
		zap.Int("user_id", saga.UserID),
		zap.String("step_name", step.Name))
	return comp, nil
}

func (h *CompensationHandler) UpdateCompensation(
	ctx context.Context,
	comp *domain.Compensation,
	status domain.CompensationStatus,
	errMsg string,
) error {
	switch status {
	case domain.CompensationStatusInProgress:
		comp.MarkInProgress()
	case domain.CompensationStatusCompleted:
		comp.MarkCompleted()
	case domain.CompensationStatusFailed:
		comp.MarkFailed(errMsg)
	}
	if err := h.repo.UpdateCompensation(ctx, comp); err != nil {
		return fmt.Errorf("update compensation: %w", err)
	}
	h.logger.Debug("compensation saved",
		zap.Int("saga_id", comp.ID),
		zap.String("user_id", comp.StepName),
		zap.String("status", status.String()))
	return nil
}

func (h *CompensationHandler) LogCompensationStart(
	saga *domain.Saga,
	step *domain.Step,
) {
	h.logger.Warn("starting compensation",
		zap.String("saga_id", saga.SagaID),
		zap.Int("user_id", saga.UserID),
		zap.String("step_name", step.Name),
		zap.String("step_status", step.Status.String()),
		zap.String("step_error", step.Error))
}

func (h *CompensationHandler) LogCompensationSuccess(
	saga *domain.Saga,
	step *domain.Step,
	duration time.Duration,
) {
	h.logger.Info("compensation completed successfully",
		zap.String("saga_id", saga.SagaID),
		zap.Int("user_id", saga.UserID),
		zap.String("step_name", step.Name),
		zap.Duration("duration", duration))
}

func (h *CompensationHandler) LogCompensationFailure(
	saga *domain.Saga,
	step *domain.Step,
	err error,
) {
	h.logger.Error("compensation failed",
		zap.String("saga_id", saga.SagaID),
		zap.Int("user_id", saga.UserID),
		zap.String("step_name", step.Name),
		zap.Error(err))
}

func (h *CompensationHandler) ShouldCompensate(step *domain.Step) bool {
	return step.IsCompleted()
}

func (h *CompensationHandler) GetCompensatableSteps(saga *domain.Saga) []*domain.Step {
	steps := make([]*domain.Step, 0)
	for i := saga.CurrentStep; i >= 0; i-- {
		step := saga.Steps[i]
		if h.ShouldCompensate(step) {
			steps = append(steps, step)
		}
	}
	return steps
}

func (h *CompensationHandler) NotifyCompensationFailure(
	ctx context.Context,
	saga *domain.Saga,
	step *domain.Step,
	err error,
) error {
	reason := fmt.Sprintf("compensation failed for step %s", step.Name)

	if pubErr := h.publisher.SendUserDeleteFailed(
		ctx,
		saga.UserID,
		saga.SagaID,
		reason,
		err.Error(),
	); pubErr != nil {
		return fmt.Errorf("publish compensation failed event: %w", pubErr)
	}

	h.logger.Warn("compensation failure event sent",
		zap.String("saga_id", saga.SagaID),
		zap.String("step_name", step.Name))

	return nil
}
