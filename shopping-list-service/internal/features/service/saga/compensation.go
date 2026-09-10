package saga

import (
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/pkg/logger"
	"go.uber.org/zap"
)

type CompensationLogger struct {
	logger *logger.Logger
}

func NewCompensationLogger(logger *logger.Logger) *CompensationLogger {
	return &CompensationLogger{
		logger: logger,
	}
}

func (c *CompensationLogger) LogCompensation(sagaID, stepName string, err error) {
	if err != nil {
		c.logger.Error("compensation failed",
			zap.String("saga_id", sagaID),
			zap.String("step_name", stepName),
			zap.Error(err))
	} else {
		c.logger.Info("compensation completed",
			zap.String("saga_id", sagaID),
			zap.String("step_name", stepName))
	}
}
