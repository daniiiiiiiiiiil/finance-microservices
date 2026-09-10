package saga

import (
	"context"
	"fmt"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/ports"
	sagacore "github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/saga"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/features/repository/postgres"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/pkg/logger"
	"go.uber.org/zap"
)

type DeleteUserSaga struct {
	logger         *logger.Logger
	sagaManager    *sagacore.SagaManager
	shoppingRepo   *postgres.ShoppingRepository
	storage        ports.StorageClient
	eventPublisher ports.EventPublisher
}

func NewDeleteUserSaga(
	logger *logger.Logger,
	sagaManager *sagacore.SagaManager,
	shoppingRepo *postgres.ShoppingRepository,
	storage ports.StorageClient,
	eventPublisher ports.EventPublisher,
) *DeleteUserSaga {
	return &DeleteUserSaga{
		logger:         logger,
		sagaManager:    sagaManager,
		shoppingRepo:   shoppingRepo,
		storage:        storage,
		eventPublisher: eventPublisher,
	}
}

func (s *DeleteUserSaga) HandleUserDeleted(ctx context.Context, userID int) error {
	s.logger.Info("starting delete user saga handler", zap.Int("user_id", userID))

	sagaID := fmt.Sprintf("delete-user-%d-%d", userID, time.Now().UnixNano())
	var shoppingIDs []int
	var imageKeys []string
	var deletedImagesCount int

	// Шаг 1 Получить все shopping пользователя
	step1 := sagacore.NewStep(
		"get_user_shoppings",
		1,
		func(ctx context.Context) error {
			shoppings, err := s.shoppingRepo.GetShoppingsByUserID(ctx, userID)
			if err != nil {
				return fmt.Errorf("get shopping: %w", err)
			}
			if len(shoppings) == 0 {
				s.logger.Info("no shoppings to delete", zap.Int("user_id", userID))
				return nil
			}
			for _, shopping := range shoppings {
				shoppingIDs = append(shoppingIDs, shopping.ID)
				if shopping.ImageKey != nil && *shopping.ImageKey != "" {
					imageKeys = append(imageKeys, *shopping.ImageKey)
				}
			}
			s.logger.Info("found shoppings to delete",
				zap.Int("user_id", userID),
				zap.Int("shopping_count", len(shoppingIDs)),
				zap.Int("image_count", len(imageKeys)))

			return nil
		},
		nil,
	)

	// Шаг 2 Удалить изображения из MinIO
	step2 := sagacore.NewStep(
		"delete_images",
		2,
		func(ctx context.Context) error {
			for _, imageKey := range imageKeys {
				if err := s.storage.Delete(ctx, imageKey); err != nil {
					s.logger.Warn("failed to delete image",
						zap.String("image_key", imageKey),
						zap.Error(err))
				} else {
					deletedImagesCount++
				}
			}
			s.logger.Info("images deleted from storage",
				zap.Int("total", len(imageKeys)),
				zap.Int("deleted", deletedImagesCount))
			return nil
		},
		nil,
	)
	// Шаг 3 Удалить shopping из БД
	step3 := sagacore.NewStep(
		"delete_shoppings",
		3,
		func(ctx context.Context) error {
			for _, shoppingID := range shoppingIDs {
				if err := s.shoppingRepo.DeleteShoppingNoTx(ctx, shoppingID, userID); err != nil {
					return fmt.Errorf("delete shopping: %w", err)
				}
			}
			s.logger.Info("shopping deleted from storage", zap.Int("shopping_count", len(shoppingIDs)))
			return nil
		},
		nil,
	)
	// Шаг 4 Отправить события в Kafka
	step4 := sagacore.NewStep(
		"send_events",
		4,
		func(ctx context.Context) error {
			if publisher, ok := s.eventPublisher.(interface {
				SendShoppingDeletedEvent(context.Context, int, int, []int, []string) error
			}); ok {
				if err := publisher.SendShoppingDeletedEvent(ctx, userID, len(shoppingIDs), shoppingIDs, imageKeys); err != nil {
					s.logger.Warn("failed to send shopping deleted event", zap.Error(err))
				}
			}
			if publisher, ok := s.eventPublisher.(interface {
				SendDeleteCompletedEvent(context.Context, int, int, int) error
			}); ok {
				if err := publisher.SendDeleteCompletedEvent(ctx, userID, len(shoppingIDs), deletedImagesCount); err != nil {
					s.logger.Warn("failed to send delete completed event", zap.Error(err))
				}
			}
			return nil
		},
		nil,
	)

	saga := sagacore.NewSaga(sagaID, "delete_user", userID, []*sagacore.Step{step1, step2, step3, step4})

	if err := s.sagaManager.StartSaga(ctx, saga); err != nil {
		s.logger.Error("delete user saga failed",
			zap.Int("user_id", userID),
			zap.Error(err))

		if publisher, ok := s.eventPublisher.(interface {
			SendDeleteFailedEvent(context.Context, int, string, error, string) error
		}); ok {
			if pubErr := publisher.SendDeleteFailedEvent(ctx, userID, "saga failed", err, "saga_execution"); pubErr != nil {
				s.logger.Error("failed to send user.delete.failed event", zap.Error(pubErr))
			}
		}
		return err
	}
	return nil
}
