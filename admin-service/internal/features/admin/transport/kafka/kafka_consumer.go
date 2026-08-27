package kafka

import (
	"context"

	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/internal/core/kafka"
	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/internal/features/admin/service"
	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/pkg/logger"
)

type AdminKafkaConsumer struct {
	consumer *kafka.Consumer
	service  *service_admin.AdminService
	logger   *logger.Logger
}

func NewAdminKafkaConsumer(
	consumer *kafka.Consumer,
	service *service_admin.AdminService,
	logger *logger.Logger,
) *AdminKafkaConsumer {
	return &AdminKafkaConsumer{
		consumer: consumer,
		service:  service,
		logger:   logger,
	}
}

func (c *AdminKafkaConsumer) Start(ctx context.Context) error {
	c.consumer.RegisterHandler(kafka.EventTypeTransactionCreated, c.service.HandleTransactionEvent)
	c.consumer.RegisterHandler(kafka.EventTypeTransactionUpdated, c.service.HandleTransactionEvent)
	c.consumer.RegisterHandler(kafka.EventTypeTransactionDeleted, c.service.HandleTransactionDeleted)
	c.consumer.RegisterHandler(kafka.EventTypeUserCreated, c.service.HandleUserCreated)
	c.consumer.RegisterHandler(kafka.EventTypeUserDeleted, c.service.HandleUserDeleted)

	return c.consumer.Start(ctx)
}

func (c *AdminKafkaConsumer) Close() error {
	return c.consumer.Close()
}
