package main

import (
	"backend/config"
	"backend/internal/core/auth/jwt"
	"backend/internal/core/cache"
	grpcclient "backend/internal/core/grpc"
	"backend/internal/core/kafka"
	"backend/internal/core/logger"
	"backend/internal/core/telemetry"
	"backend/internal/core/transport/grpc/interceptors"
	"backend/internal/features/auth/repository/redis"
	redisCache "backend/internal/features/currency/repository/redis"
	service_currency "backend/internal/features/currency/service"
	currency_grpc "backend/internal/features/currency/transport/grpc"
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	cfg := config.NewConfigMust()
	time.Local = cfg.TimeZone

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
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

	jwtManager := jwt.NewJWTManager(cfg.JWTSecret, cfg.JWTDuration)
	blacklist := redis.NewBlacklistCache(redisClient)

	logger.Debug("initializing kafka producer")
	kafkaConfig := kafka.NewConfig()
	kafkaProducer := kafka.NewProducer(kafkaConfig, *logger)
	defer kafkaProducer.Close()

	serviceName := "currency"
	shutdown, err := telemetry.InitTracer(serviceName)
	if err != nil {
		logger.Fatal("failed to init tracer", zap.Error(err))
	}
	defer shutdown()

	apiURL := os.Getenv("CURRENCY_API_URL")
	if apiURL == "" {
		apiURL = "https://api.frankfurter.app/latest"
	}
	currencyClient := service_currency.NewCurrencyClient(apiURL, *logger)

	rateCache := redisCache.NewRateCache(redisClient)

	currencyService := service_currency.NewCurrencyService(
		rateCache,
		currencyClient,
		logger,
		redisClient,
		kafkaProducer,
	)

	go currencyService.StartConsumer(ctx)

	logger.Debug("initializing currency grpc server")
	grpcServer := grpcclient.NewGRPCServer(
		cfg,
		interceptors.RequestIDInterceptor(),
		interceptors.LoggerInterceptor(logger),
		interceptors.AuthInterceptor(jwtManager, blacklist),
		interceptors.MetricsInterceptor(serviceName),
		interceptors.TraceInterceptor(),
	)

	currencyServer := currency_grpc.NewCurrencyServer(currencyService, logger)
	currency_grpc.RegisterCurrencyServer(grpcServer, currencyServer)

	metricsPort := ":9095"
	metricsServer := &http.Server{Addr: metricsPort, Handler: promhttp.Handler()}
	go func() {
		logger.Info("starting metrics server", zap.String("addr", metricsPort))
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("metrics server failed", zap.Error(err))
		}
	}()

	lis, err := net.Listen("tcp", ":50055")
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	logger.Warn("starting gRPC server", zap.String("address", ":50055"), zap.String("service", "currency"))
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
