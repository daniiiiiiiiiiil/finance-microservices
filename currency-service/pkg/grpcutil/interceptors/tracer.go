package interceptors

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc/stats"
)

func TraceInterceptor() stats.Handler {
	return otelgrpc.NewServerHandler()
}
