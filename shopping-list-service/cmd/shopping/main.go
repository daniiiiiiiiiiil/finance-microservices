package shopping

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/config"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/auth/jwt"
	grpcclient "github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/grpc"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/repository/postgres/pool/pgx"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/telemetry"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/features/repository/postgres"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/features/service"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/features/transport/gRPC"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/pkg/grpcutil/interceptors"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/pkg/logger"
	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/proto/shopping/gen"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	cfg := config.NewConfigMust()
	time.Local = cfg.TimeZone

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger, err := logger.NewLogger(logger.NewConfigMust())
	if err != nil {
		os.Exit(1)
	}
	defer logger.Close()

	logger.Debug("application time zone", zap.Any("time_zone", time.Local))

	logger.Debug("initializing postgres connection pool")
	pool, err := pgx.NewPool(ctx, pgx.NewConfigMust())
	if err != nil {
		logger.Fatal("postgres connection error", zap.Error(err))
	}
	defer pool.Close()

	logger.Debug("initializing jwt shopping service")
	jwtManager := jwt.NewJWTManager(cfg.JWTSecret, cfg.JWTDuration)
	serviceName := "shopping-list"
	shutdown, err := telemetry.InitTracer(serviceName)
	if err != nil {
		logger.Fatal("telemetry error", zap.Error(err))
	}
	defer shutdown()

	logger.Debug("initializing shopping server postgres")
	shoppingRepository := postgres.NewShoppingRepository(pool)

	shoppingService := service.NewShoppingService(
		shoppingRepository,
		pool,
		logger,
	)

	logger.Debug("initializing shopping service")
	grpcServer := grpcclient.NewGRPCServer(
		cfg,
		interceptors.RequestIDInterceptor(),
		interceptors.LoggerInterceptor(logger),
		interceptors.AuthInterceptor(jwtManager),
		interceptors.MetricsInterceptor(serviceName),
		interceptors.TraceInterceptor())

	shoppingServer := gRPC.NewShoppingListService(shoppingService, logger)
	gen.RegisterShoppingServiceServer(grpcServer, shoppingServer)

	metricsPort := ":9097"
	metricsServer := http.Server{
		Addr:    metricsPort,
		Handler: promhttp.Handler(),
	}
	go func() {
		logger.Info("starting metrics server", zap.String("port", metricsPort))
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("metrics server error", zap.Error(err))
		}
	}()

	lis, err := net.Listen("tcp", ":50060")
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	logger.Warn("starting grpc server", zap.String("port", metricsPort))
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("grpc server error", zap.Error(err))
		}
	}()

	<-ctx.Done()

	logger.Warn("shutting down grpc server")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	done := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("grpc server gracefully stopped")
	case <-shutdownCtx.Done():
		logger.Warn("grpc server shutdown timed out stopped")
		grpcServer.Stop()
	}
	if err := metricsServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("metrics server error", zap.Error(err))
	}
	logger.Info("shutdown complete")
}
