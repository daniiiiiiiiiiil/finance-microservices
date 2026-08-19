package main

import (
	"backend/internal/core/auth/jwt"
	"backend/internal/core/cache"
	"backend/internal/core/config"
	"backend/internal/core/logger"
	"backend/internal/core/repository/postgres/pool/pgx"
	"backend/internal/core/transport/grpc/interceptors"
	"backend/internal/features/auth/repository/redis"
	"backend/internal/features/users/repository/postgres"
	"backend/internal/features/users/service"
	usersgrpc "backend/internal/features/users/transport/grpc"

	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.NewConfigMust()

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

	logger.Debug("initializing redis")
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisClient := cache.NewRedisClient(redisAddr)
	defer redisClient.Close()

	logger.Debug("application time zone", zap.Any("time_zone", time.Local))

	logger.Debug("initializing postgres connection pool")
	pool, err := pgx.NewPool(ctx, pgx.NewConfigMust())
	if err != nil {
		logger.Fatal("failed to connect to postgres", zap.Error(err))
	}
	defer pool.Close()

	jwtManager := jwt.NewJWTManager(cfg.JWTSecret, cfg.JWTDuration)
	logger.Debug("JWT Manager initialized")

	blacklist := redis.NewBlacklistCache(redisClient)

	logger.Debug("initializing users service")
	usersRepository := postgres.NewUserRepository(pool)
	usersService := service_user.NewUsersService(usersRepository, pool, redisClient)

	logger.Debug("initializing gRPC server with interceptors")

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.RequestIDInterceptor(),
			interceptors.LoggerInterceptor(logger),
			interceptors.AuthInterceptor(jwtManager, blacklist),
			interceptors.TraceInterceptor(logger.Logger)),
	)

	userServer := usersgrpc.NewUserServer(usersService, logger)
	usersgrpc.RegisterUserServer(grpcServer, userServer)

	reflection.Register(grpcServer)

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
	logger.Warn("shutting down gRPC server", zap.String("address", ":50052"))
	grpcServer.GracefulStop()
	logger.Warn("gRPC server stopped", zap.String("address", ":50052"))
}
