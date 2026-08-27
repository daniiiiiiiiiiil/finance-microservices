package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/config"
	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/core/auth/jwt"
	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/core/cache"
	usersclient "github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/core/clients/users"
	grpcclient "github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/core/grpc"
	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/core/repository/postgres/pool/pgx"
	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/core/telemetry"
	postgres_auth "github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/features/auth/repository/postgres"
	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/features/auth/repository/redis"
	service_auth "github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/features/auth/service"
	authgrpc "github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/features/auth/transport/grpc"
	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/pkg/grpcutil/interceptors"
	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/pkg/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc/reflection"
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
		redisAddr = "finance-redis:6379"
	}
	redisClient := cache.NewRedisClient(redisAddr)
	defer redisClient.Close()

	jwtManager := jwt.NewJWTManager(cfg.JWTSecret, cfg.JWTDuration)
	blacklist := redis.NewBlacklistCache(redisClient)

	serviceName := "auth"
	shutdown, err := telemetry.InitTracer(serviceName)
	if err != nil {
		logger.Fatal("failed to init tracer", zap.Error(err))
	}
	defer shutdown()

	logger.Debug("initializing users client")
	usersClient, err := usersclient.NewUsersClient("users:50052", cfg)
	if err != nil {
		logger.Fatal("failed to connect to users service", zap.Error(err))
	}
	defer usersClient.Close()

	logger.Debug("initializing auth service")
	authRepo := postgres_auth.NewAuthRepository(pool)
	authService := service_auth.NewAuthService(authRepo, jwtManager, redisClient, usersClient)
	createFirstAdmin(ctx, authService, logger)

	logger.Debug("initializing auth service gRPC")
	grpcServer := grpcclient.NewGRPCServer(
		cfg,
		interceptors.RequestIDInterceptor(),
		interceptors.LoggerInterceptor(logger),
		interceptors.AuthInterceptor(jwtManager, blacklist),
		interceptors.MetricsInterceptor(serviceName),
		interceptors.TraceInterceptor(),
	)

	authServer := authgrpc.NewAuthServer(authService, logger)
	authgrpc.RegisterAuthServer(grpcServer, authServer)

	reflection.Register(grpcServer)

	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())

	metricsPort := ":9091"
	metricsServer := &http.Server{
		Addr:    metricsPort,
		Handler: metricsMux,
	}

	go func() {
		logger.Info("starting metrics server", zap.String("addr", metricsPort))
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("metrics server failed", zap.Error(err))
		}
	}()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	logger.Warn("starting gRPC server", zap.String("address", ":50051"), zap.String("service", "auth"))

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

func createFirstAdmin(ctx context.Context, authService *service_auth.AuthService, log *logger.Logger) {
	req := ports.RegisterRequest{
		FullName:    "Admin",
		Email:       "admin@finance.com",
		Password:    "admin123",
		PhoneNumber: "+79998887766",
		IsAdmin:     true,
	}

	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		token, user, err := authService.Register(ctx, req)
		if err == nil {
			log.Info("first admin created successfully",
				zap.String("email", user.Email),
				zap.String("token", token[:20]+"..."),
			)
			return
		}

		if strings.Contains(err.Error(), "already exists") {
			log.Info("admin already exists, skipping creation")
			return
		}

		log.Warn("failed to create first admin, retrying...",
			zap.Error(err),
			zap.Int("attempt", i+1),
		)
		time.Sleep(2 * time.Second)
	}
}
