package main

import (
	"backend/internal/core/auth/jwt"
	"backend/internal/core/cache"
	"backend/internal/core/config"
	"backend/internal/core/kafka"
	"backend/internal/core/logger"
	"backend/internal/core/repository/postgres/pool/pgx"
	core_http_middleware "backend/internal/core/transport/http/middleware"
	"backend/internal/core/transport/http/server"
	admin_repo "backend/internal/features/admin/repository/postgres"
	admin_service "backend/internal/features/admin/service"
	admin_http "backend/internal/features/admin/transport/http"
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	_ "backend/docs"
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

	jwtManager := jwt.NewJWTManager(cfg.JWTSecret, cfg.JWTDuration)

	logger.Debug("initializing admin service")
	adminRepository := admin_repo.NewAdminRepository(pool)
	adminService := admin_service.NewAdminService(adminRepository, pool, redisClient, kafkaProducer)
	adminHandler := admin_http.NewAdminHandler(adminService, jwtManager)

	logger.Debug("initializing http server")
	httpServer := server.NewHTTPServer(
		server.NewConfigMust(),
		logger,
		core_http_middleware.CORS(),
		core_http_middleware.RequestID(),
		core_http_middleware.Logger(logger),
		core_http_middleware.Trace(),
		core_http_middleware.Panic(),
	)

	httpServer.RegisterRoutes(adminHandler.Routes()...)

	if err := httpServer.Run(ctx); err != nil {
		logger.Error("Failed to start server", zap.Error(err))
	}
}
