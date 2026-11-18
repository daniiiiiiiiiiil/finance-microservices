package interceptors

import (
	"context"

	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
)

func TraceInterceptor() grpc.UnaryServerInterceptor {
	tracer := otel.Tracer("telegram-bot-service")
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx, span := tracer.Start(ctx, info.FullMethod)
		defer span.End()
		return handler(ctx, req)
	}
}
