package interceptors

import (
	"context"
	"strings"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func MetricsInterceptor(serviceName string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		method := info.FullMethod
		if index := strings.LastIndex(method, "/"); index >= 0 && index < len(method)-1 {
			method = method[index+1:]
		}
		finish := metrics.StartGrpcRequest(serviceName)
		defer finish()

		start := time.Now()
		resp, err := handler(ctx, req)
		metrics.RecordGrpcRequest(serviceName, method, status.Code(err).String(), time.Since(start))
		return resp, err
	}
}
