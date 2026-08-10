package interceptors

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func TraceInterceptor(log *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		log.Debug("grpc request completed",
			zap.String("method", info.FullMethod),
			zap.Duration("elapsed", time.Since(start)),
			zap.Error(err),
		)

		return resp, err
	}
}
