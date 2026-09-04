package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc/reflection"

	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/config"
	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/auth/jwt"
	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/cache"
	grpcclient "github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/grpc"
	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/kafka"
	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/repository/postgres/pool/pgx"
	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/core/telemetry"
	postgres_users "github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/features/users/repository/postgres"
	redis_cache "github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/features/users/repository/redis"
	service_user "github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/features/users/service"
	usersgrpc "github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/features/users/transport/grpc"
	transportkafka "github.com/daniiiiiiiiiiil/finance-microservices/users-service/internal/features/users/transport/kafka"
	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/pkg/grpcutil/interceptors"
	"github.com/daniiiiiiiiiiil/finance-microservices/users-service/pkg/logger"
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

	logger.Debug("initializing postgres connection pool")
	pool, err := pgx.NewPool(ctx, pgx.NewConfigMust())
	if err != nil {
		logger.Fatal("failed to connect to postgres", zap.Error(err))
	}
	defer pool.Close()

	jwtManager := jwt.NewJWTManager(cfg.JWTSecret, cfg.JWTDuration)
	logger.Debug("JWT Manager initialized")

	logger.Debug("initializing kafka producer")
	kafkaConfig := kafka.NewConfig()
	kafkaProducer := kafka.NewProducer(kafkaConfig, *logger)
	defer kafkaProducer.Close()

	eventPublisher := transportkafka.NewUserEventPublisher(kafkaProducer)

	serviceName := "users"
	shutdown, err := telemetry.InitTracer(serviceName)
	if err != nil {
		logger.Fatal("failed to init tracer", zap.Error(err))
	}
	defer shutdown()

	logger.Debug("initializing users repository and caches")
	usersRepository := postgres_users.NewUserRepository(pool)
	userCache := redis_cache.NewUserCache(redisClient)
	usersListCache := redis_cache.NewUsersListCache(redisClient)

	logger.Debug("initializing outbox")
	outboxRepository := postgres_users.NewOutboxRepository(pool)
	outboxPublisher := service_user.NewOutboxPublisher(outboxRepository, kafkaProducer, logger)
	outboxPublisher.Start(ctx)

	logger.Debug("initializing users service")
	usersService := service_user.NewUsersService(
		usersRepository,
		pool,
		userCache,
		usersListCache,
		eventPublisher,
		outboxRepository,
		logger,
		redisClient,
	)

	logger.Debug("initializing gRPC server with interceptors")

	grpcServer := grpcclient.NewGRPCServer(
		cfg,
		interceptors.RequestIDInterceptor(),
		interceptors.LoggerInterceptor(logger),
		interceptors.AuthInterceptor(jwtManager),
		interceptors.MetricsInterceptor(serviceName),
		interceptors.TraceInterceptor(),
	)
	userServer := usersgrpc.NewUserServer(usersService, logger)
	usersgrpc.RegisterUserServer(grpcServer, userServer)

	reflection.Register(grpcServer)

	metricsPort := ":9092"
	metricsServer := &http.Server{Addr: metricsPort, Handler: promhttp.Handler()}
	go func() {
		logger.Info("starting metrics server", zap.String("addr", metricsPort))
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("metrics server failed", zap.Error(err))
		}
	}()

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	logger.Warn("starting gRPC server",
		zap.String("address", ":50052"),
		zap.String("service", "users"),
	)

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
