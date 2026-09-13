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

type RegisterUserSaga struct {
	logger      *logger.Logger
	sagaManager ports.SagaManager
	publisher   ports.EventPublisher
	repo        ports.SagaRepository
}

func NewRegisterUserSaga(
	logger *logger.Logger,
	sagaManager ports.SagaManager,
	publisher ports.EventPublisher,
	repo ports.SagaRepository,
) *RegisterUserSaga {
	return &RegisterUserSaga{
		logger:      logger,
		sagaManager: sagaManager,
		publisher:   publisher,
		repo:        repo,
	}
}

func (s *RegisterUserSaga) Definition() ports.SagaDefinition {
	return ports.SagaDefinition{
		Type: domain.SagaTypeRegisterUser,
		Steps: []ports.SagaStepDefinition{
			{
				Name:       "create_credentials",
				Order:      1,
				Action:     s.createCredentials,
				Compensate: s.compensateCreateCredentials,
			},
			{
				Name:       "create_profile",
				Order:      2,
				Action:     s.createProfile,
				Compensate: s.compensateCreateProfile,
			},
			{
				Name:       "activate_account",
				Order:      3,
				Action:     s.activateAccount,
				Compensate: nil,
			},
		},
		Build: s.build,
	}
}

func (s *RegisterUserSaga) HandleRegisterRequested(
	ctx context.Context,
	email string,
	fullName string,
	passwordHash string,
	phoneNumber *string,
	isAdmin bool,
) error {
	s.logger.Info("handling register requested",
		zap.String("email", email),
		zap.String("full_name", fullName),
		zap.Bool("is_admin", isAdmin))

	if email == "" {
		return fmt.Errorf("email is required")
	}
	if fullName == "" {
		return fmt.Errorf("full_name is required")
	}
	if passwordHash == "" {
		return fmt.Errorf("password_hash is required")
	}

	sagaID := fmt.Sprintf("register-user-%s-%d", email, time.Now().UnixNano())
	totalSteps := 3

	saga := domain.NewSaga(sagaID, domain.SagaTypeRegisterUser, 0, totalSteps)
	saga.SetMetadata("email", email)
	saga.SetMetadata("full_name", fullName)
	saga.SetMetadata("password_hash", passwordHash)
	saga.SetMetadata("is_admin", isAdmin)
	if phoneNumber != nil {
		saga.SetMetadata("phone_number", *phoneNumber)
	}
	saga.SetMetadata("started_at", time.Now().Format(time.RFC3339))

	if err := s.sagaManager.StartSaga(ctx, saga); err != nil {
		s.logger.Error("failed to start register user saga",
			zap.String("saga_id", sagaID),
			zap.String("email", email),
			zap.Error(err))
		return fmt.Errorf("start saga: %w", err)
	}

	s.logger.Info("register user saga completed",
		zap.String("saga_id", sagaID),
		zap.String("email", email))

	return nil
}

func (s *RegisterUserSaga) build(ctx context.Context, saga *domain.Saga) error {
	steps := []struct {
		name  string
		order int
	}{
		{"create_credentials", 1},
		{"create_profile", 2},
		{"activate_account", 3},
	}

	for _, step := range steps {
		saga.AddStep(domain.NewStep(saga.ID, step.name, step.order))
	}

	return nil
}

func (s *RegisterUserSaga) createCredentials(ctx context.Context, saga *domain.Saga) error {
	s.logger.Info("step: create_credentials",
		zap.String("saga_id", saga.SagaID))

	email, _ := saga.GetMetadata("email")
	passwordHash, _ := saga.GetMetadata("password_hash")

	if err := s.publisher.Publish(ctx, "auth.create_credentials", map[string]interface{}{
		"saga_id":       saga.SagaID,
		"email":         email,
		"password_hash": passwordHash,
	}); err != nil {
		return fmt.Errorf("send auth.create_credentials: %w", err)
	}

	s.logger.Info("auth.create_credentials command sent",
		zap.String("saga_id", saga.SagaID))

	return nil
}

func (s *RegisterUserSaga) compensateCreateCredentials(ctx context.Context, saga *domain.Saga) error {
	s.logger.Warn("compensating: create_credentials",
		zap.String("saga_id", saga.SagaID))

	email, _ := saga.GetMetadata("email")

	if err := s.publisher.Publish(ctx, "auth.delete_credentials", map[string]interface{}{
		"saga_id": saga.SagaID,
		"email":   email,
	}); err != nil {
		return fmt.Errorf("send auth.delete_credentials: %w", err)
	}

	return nil
}

func (s *RegisterUserSaga) createProfile(ctx context.Context, saga *domain.Saga) error {
	s.logger.Info("step: create_profile",
		zap.String("saga_id", saga.SagaID))

	email, _ := saga.GetMetadata("email")
	fullName, _ := saga.GetMetadata("full_name")
	isAdmin, _ := saga.GetMetadata("is_admin")
	phoneNumber, _ := saga.GetMetadata("phone_number")

	if err := s.publisher.Publish(ctx, "user.create_profile", map[string]interface{}{
		"saga_id":      saga.SagaID,
		"email":        email,
		"full_name":    fullName,
		"is_admin":     isAdmin,
		"phone_number": phoneNumber,
	}); err != nil {
		return fmt.Errorf("send user.create_profile: %w", err)
	}

	s.logger.Info("user.create_profile command sent",
		zap.String("saga_id", saga.SagaID))

	return nil
}

func (s *RegisterUserSaga) compensateCreateProfile(ctx context.Context, saga *domain.Saga) error {
	s.logger.Warn("compensating: create_profile",
		zap.String("saga_id", saga.SagaID))

	email, _ := saga.GetMetadata("email")

	if err := s.publisher.Publish(ctx, "user.delete_profile", map[string]interface{}{
		"saga_id": saga.SagaID,
		"email":   email,
	}); err != nil {
		return fmt.Errorf("send user.delete_profile: %w", err)
	}

	return nil
}

func (s *RegisterUserSaga) activateAccount(ctx context.Context, saga *domain.Saga) error {
	s.logger.Info("step: activate_account",
		zap.String("saga_id", saga.SagaID))

	email, _ := saga.GetMetadata("email")

	if err := s.publisher.Publish(ctx, "auth.activate_account", map[string]interface{}{
		"saga_id": saga.SagaID,
		"email":   email,
	}); err != nil {
		return fmt.Errorf("send auth.activate_account: %w", err)
	}

	s.logger.Info("register user saga completed",
		zap.String("saga_id", saga.SagaID))

	return nil
}
