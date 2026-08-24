package main

import (
	"backend/config"
	"backend/internal/core/auth/jwt"
	"backend/internal/core/cache"
	grpcclient "backend/internal/core/grpc"
	"backend/internal/core/kafka"
	"backend/internal/core/repository/postgres/pool/pgx"
	"backend/internal/core/telemetry"
	"backend/internal/features/auth/repository/redis"
	finance_repo "backend/internal/features/finance/repository/postgres"
	finance_service "backend/internal/features/finance/service"
	finance_grpc "backend/internal/features/finance/transport/grpc"
	interceptors2 "backend/pkg/grpcutil/interceptors"
	logger2 "backend/pkg/logger"
	"backend/proto/finance/gen"
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "backend/docs"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	cfg := config.NewConfigMust()
	time.Local = cfg.TimeZone

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM)
	defer cancel()

	logger, err := logger2.NewLogger(logger2.NewConfigMust())
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

	jwtManager := jwt.NewJWTManager(cfg.JWTSecret, cfg.JWTDuration)

	blacklist := redis.NewBlacklistCache(redisClient)

	serviceName := "finance"
	shutdown, err := telemetry.InitTracer(serviceName)
	if err != nil {
		logger.Fatal("failed to init tracer", zap.Error(err))
	}
	defer shutdown()

	logger.Debug("initializing finance service")
	financeRepository := finance_repo.NewFinanceRepository(pool)
	financeService := finance_service.NewFinanceService(financeRepository, pool, redisClient, kafkaProducer, logger)

	logger.Debug("initializing finance grpc server")
	grpcServer := grpcclient.NewGRPCServer(
		cfg,
		interceptors2.RequestIDInterceptor(),
		interceptors2.LoggerInterceptor(logger),
		interceptors2.AuthInterceptor(jwtManager, blacklist),
		interceptors2.MetricsInterceptor(serviceName),
		interceptors2.TraceInterceptor(),
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
