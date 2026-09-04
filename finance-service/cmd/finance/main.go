package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/s3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/config"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/auth/jwt"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/cache"
	grpcclient "github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/grpc"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/kafka"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/repository/postgres/pool/pgx"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/core/telemetry"
	finance_repo "github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/features/finance/repository/postgres"
	finance_service "github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/features/finance/service"
	finance_grpc "github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/features/finance/transport/grpc"
	transportkafka "github.com/daniiiiiiiiiiil/finance-microservices/finance-service/internal/features/finance/transport/kafka"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/pkg/grpcutil/interceptors"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/pkg/logger"
	"github.com/daniiiiiiiiiiil/finance-microservices/finance-service/proto/finance/gen"
)

func main() {
	cfg := config.NewConfigMust()
	time.Local = cfg.TimeZone

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM)
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
		logger.Fatal("failed to connect to postgres", zap.Error(err))
	}
	defer pool.Close()

	logger.Debug("initializing redis")
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisClient := cache.NewRedisClient(redisAddr)
	defer redisClient.Close()

	logger.Debug("initializing kafka producer")
	kafkaConfig := kafka.NewConfig()
	kafkaProducer := kafka.NewProducer(kafkaConfig, *logger)
	defer kafkaProducer.Close()

	eventPublisher := transportkafka.NewFinanceEventPublisher(kafkaProducer)

	jwtManager := jwt.NewJWTManager(cfg.JWTSecret, cfg.JWTDuration)

	serviceName := "finance"
	shutdown, err := telemetry.InitTracer(serviceName)
	if err != nil {
		logger.Fatal("failed to init tracer", zap.Error(err))
	}
	defer shutdown()

	logger.Debug("initializing finance service")
	financeRepository := finance_repo.NewFinanceRepository(pool)

	logger.Debug("initializing S3 client")
	s3Client, err := s3.NewClient()
	if err != nil {
		logger.Fatal("failed to create S3 client", zap.Error(err))
	}
	logger.Debug("initializing export service")
	exportService := finance_service.NewExportService(financeRepository, s3Client)

	logger.Debug("initializing outbox")
	outboxRepository := finance_repo.NewOutboxRepository(pool)

	financeService := finance_service.NewFinanceService(
		financeRepository,
		pool,
		redisClient,
		eventPublisher,
		outboxRepository,
		exportService,
		logger,
	)

	outboxPublisher := finance_service.NewOutboxPublisher(outboxRepository, kafkaProducer, logger)
	outboxPublisher.Start(ctx)

	logger.Debug("initializing finance grpc server")
	grpcServer := grpcclient.NewGRPCServer(
		cfg,
		interceptors.RequestIDInterceptor(),
		interceptors.LoggerInterceptor(logger),
		interceptors.AuthInterceptor(jwtManager),
		interceptors.MetricsInterceptor(serviceName),
		interceptors.TraceInterceptor(),
	)

	financeServer := finance_grpc.NewFinanceServer(financeService, logger)
	gen.RegisterFinanceServiceServer(grpcServer, financeServer)

	metricsPort := ":9093"
	metricsServer := &http.Server{Addr: metricsPort, Handler: promhttp.Handler()}
	go func() {
		logger.Info("starting metrics server", zap.String("addr", metricsPort))
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("metrics server failed", zap.Error(err))
		}
	}()

	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	logger.Warn("starting gRPC server", zap.String("address", ":50053"), zap.String("service", "finance"))
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("failed to serve", zap.Error(err))
		}
	}()

	<-ctx.Done()

	logger.Warn("shutting down gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	done := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("gRPC server stopped gracefully")
	case <-shutdownCtx.Done():
		logger.Warn("gRPC server stop timeout, forcing stop")
		grpcServer.Stop()
	}

	if err := metricsServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("failed to shutdown metrics server", zap.Error(err))
	}

	logger.Info("shutdown complete")
}
