package interceptors

import (
	"strings"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/core/metrics"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func MetricsInterceptor(serviceName string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		method := info.FullMethod
		parts := strings.Split(method, "/")
		if len(parts) > 0 {
			method = parts[len(parts)-1]
		}
		metrics.RecordGrpcRequestStart(serviceName)
		defer metrics.RecordGrpcRequestFinish(serviceName)

		start := time.Now()
		resp, err := handler(ctx, req)

		duration := time.Since(start).Seconds()
		st := status.Code(err).String()

		metrics.RecordGrpcRequest(serviceName, method, st, duration)
		return resp, err
	}
}
