package orchestrator

import (
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/pkg/logger"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type DeleteUserSaga struct {
	logger      *logger.Logger
	sagaManager ports.SagaManager
	publisher   ports.EventPublisher
	repo        ports.SagaRepository
}

func NewDeleteUserSaga(
	logger *logger.Logger,
	sagaManager ports.SagaManager,
	publisher ports.EventPublisher,
	repo ports.SagaRepository,
) *DeleteUserSaga {
	return &DeleteUserSaga{
		logger:      logger,
		sagaManager: sagaManager,
		publisher:   publisher,
		repo:        repo,
	}
}

func (s *DeleteUserSaga) Definition() ports.SagaDefinition {
	return ports.SagaDefinition{
		Type: domain.SagaTypeDeleteUser,
		Steps: []ports.SagaStepDefinition{
			{
				Name:       "mark_deleting",
				Order:      1,
				Action:     s.markDeleting,
				Compensate: s.compensateMarkDeleting,
			},
			{
				Name:       "delete_shopping",
				Order:      2,
				Action:     s.deleteShopping,
				Compensate: nil,
			},
			{
				Name:       "delete_transactions",
				Order:      3,
				Action:     s.deleteTransactions,
				Compensate: nil,
			},
			{
				Name:       "finalize_delete",
				Order:      4,
				Action:     s.finalizeDelete,
				Compensate: nil,
			},
		},
		Build: s.build,
	}
}

func (s *DeleteUserSaga) HandleUserDeleted(ctx context.Context, userID int) error {
	s.logger.Info("Handling User Deleted Saga", zap.Int("user_id", userID))

	sagaID := fmt.Sprintf("delete-user-%d-%d", userID, time.Now().UnixNano())
	totalSteps := 4

	saga := domain.NewSaga(sagaID, domain.SagaTypeDeleteUser, userID, totalSteps)
	saga.SetMetadata("user_id", userID)
	saga.SetMetadata("started_at", time.Now().Format(time.RFC3339))

	if err := s.sagaManager.StartSaga(ctx, saga); err != nil {
		s.logger.Error("failed to start delete user saga",
			zap.String("saga_id", sagaID),
			zap.Int("user_id", userID),
			zap.Error(err))
		return fmt.Errorf("start saga: %w", err)
	}

	s.logger.Info("delete user saga completed successfully",
		zap.String("saga_id", sagaID),
		zap.Int("user_id", userID))

	return nil
}

func (s *DeleteUserSaga) build(ctx context.Context, saga *domain.Saga) error {
	steps := []struct {
		name  string
		order int
	}{
		{"mark_deleting", 1},
		{"delete_shopping", 2},
		{"delete_transactions", 3},
		{"finalize_delete", 4},
	}

	for _, step := range steps {
		saga.AddStep(domain.NewStep(saga.ID, step.name, step.order))
	}

	return nil
}

func (s *DeleteUserSaga) markDeleting(ctx context.Context, saga *domain.Saga) error {
	s.logger.Info("step: mark_deleting",
		zap.String("saga_id", saga.SagaID),
		zap.Int("user_id", saga.UserID))

	if err := s.publisher.SendUserMarkDeleting(ctx, saga.UserID, saga.SagaID); err != nil {
		return fmt.Errorf("mark deleting: %w", err)
	}
	s.logger.Info("step: mark_deleting completed successfully", zap.String("saga_id", saga.SagaID), zap.Int("user_id", saga.UserID))
	return nil
}

func (s *DeleteUserSaga) compensateMarkDeleting(ctx context.Context, saga *domain.Saga) error {
	s.logger.Warn("compensating: mark_deleting",
		zap.String("saga_id", saga.SagaID),
		zap.Int("user_id", saga.UserID))

	if err := s.publisher.SendUserRestore(ctx, saga.UserID, saga.SagaID); err != nil {
		s.logger.Error("failed to send user.restore command",
			zap.String("saga_id", saga.SagaID),
			zap.Int("user_id", saga.UserID),
			zap.Error(err))
		return fmt.Errorf("send user.restore: %w", err)
	}

	s.logger.Info("user.restore command sent",
		zap.String("saga_id", saga.SagaID),
		zap.Int("user_id", saga.UserID))

	return nil
}

func (s *DeleteUserSaga) deleteShopping(ctx context.Context, saga *domain.Saga) error {
	s.logger.Info("step: delete_shopping",
		zap.String("saga_id", saga.SagaID),
		zap.Int("user_id", saga.UserID))

	if err := s.publisher.SendDeleteShoppingData(ctx, saga.UserID, saga.SagaID); err != nil {
		return fmt.Errorf("delete shopping: %w", err)
	}
	s.logger.Info("shopping.delete_user_data command sent",
		zap.String("saga_id", saga.SagaID),
		zap.Int("user_id", saga.UserID))

	return nil
}

func (s *DeleteUserSaga) deleteTransactions(ctx context.Context, saga *domain.Saga) error {
	s.logger.Info("step: delete_transactions",
		zap.String("saga_id", saga.SagaID),
		zap.Int("user_id", saga.UserID))

	if err := s.publisher.SendDeleteFinanceTransactions(ctx, saga.UserID, saga.SagaID); err != nil {
		return fmt.Errorf("delete transactions: %w", err)
	}
	s.logger.Info("finance.delete_user_transactions command sent",
		zap.String("saga_id", saga.SagaID),
		zap.Int("user_id", saga.UserID))

	return nil
}

func (s *DeleteUserSaga) finalizeDelete(ctx context.Context, saga *domain.Saga) error {
	s.logger.Info("step: finalize_delete",
		zap.String("saga_id", saga.SagaID),
		zap.Int("user_id", saga.UserID))

	if err := s.publisher.SendUserFinalizeDelete(ctx, saga.UserID, saga.SagaID); err != nil {
		return fmt.Errorf("send user.finalize_delete: %w", err)
	}

	s.logger.Info("user.finalize_delete command sent",
		zap.String("saga_id", saga.SagaID),
		zap.Int("user_id", saga.UserID))

	return nil
}
