package kafka

import (
	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/ports"
)

var _ ports.EventPublisher = (*SagaEventPublisher)(nil)

var _ ports.EventConsumer = (*SagaKafkaConsumer)(nil)
