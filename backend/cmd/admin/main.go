package main

import (
	"backend/config"
	"backend/internal/core/auth/jwt"
	"backend/internal/core/cache"
	financeСlient "backend/internal/core/clients/finance"
	usersclient "backend/internal/core/clients/users"
	grpcclient "backend/internal/core/grpc"
	"backend/internal/core/kafka"
	"backend/internal/core/logger"
	"backend/internal/core/telemetry"
	"backend/internal/core/transport/grpc/interceptors"
	admin_service "backend/internal/features/admin/service"
	admingrpc "backend/internal/features/admin/transport/grpc"
	"backend/internal/features/auth/repository/redis"
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

	logger, err := logger.NewLogger(logger.NewConfigMust())
	if err != nil {
		os.Exit(1)
	}
	defer logger.Close()

	logger.Debug("application time zone", zap.Any("time_zone", time.Local))

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

	serviceName := "admin"
	shutdown, err := telemetry.InitTracer(serviceName)
	if err != nil {
		logger.Fatal("failed to init tracer", zap.Error(err))
	}
	defer shutdown()

	logger.Debug("Initializing user client")
	usersClient, err := usersclient.NewUsersClient("users:50052", cfg)
	if err != nil {
		logger.Fatal("failed to initialize user client", zap.Error(err))
	}
	defer usersClient.Close()

	financeClient, err := financeСlient.NewFinanceClient("finance:50053", cfg)
	if err != nil {
		logger.Fatal("failed to connect to finance service", zap.Error(err))
	}
	defer financeClient.Close()

	logger.Debug("initializing admin service")
	adminService := admin_service.NewAdminService(
		redisClient,
		kafkaProducer,
		usersClient,
		financeClient,
		logger)
	grpcServer := grpcclient.NewGRPCServer(
		cfg,
		interceptors.RequestIDInterceptor(),
		interceptors.LoggerInterceptor(logger),
		interceptors.AuthInterceptor(jwtManager, blacklist),
		interceptors.MetricsInterceptor(serviceName),
		interceptors.TraceInterceptor(),
	)
	adminServer := admingrpc.NewAdminServer(adminService, logger)
	admingrpc.RegisterAdminServer(grpcServer, adminServer)

	metricsPort := ":9094"
	metricsServer := &http.Server{Addr: metricsPort, Handler: promhttp.Handler()}
	go func() {
		logger.Info("starting metrics server", zap.String("addr", metricsPort))
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("metrics server failed", zap.Error(err))
		}
	}()

	lis, err := net.Listen("tcp", ":50054")
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	logger.Warn("starting gRPC server", zap.String("address", ":50054"), zap.String("service", "admin"))

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("failed to serve gRPC", zap.Error(err))
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
