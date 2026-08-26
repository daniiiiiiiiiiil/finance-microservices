package interceptors

import (
	"context"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/currency-service/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func LoggerInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		ctx = logger.ToContext(ctx, log)

		log.Debug("incoming grpc request",
			zap.String("method", info.FullMethod),
		)

		resp, err := handler(ctx, req)

		log.Debug("done grpc request",
			zap.String("method", info.FullMethod),
			zap.Duration("elapsed", time.Since(start)),
			zap.Error(err),
		)

		return resp, err
	}
}
