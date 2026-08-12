package main

import (
	"backend/internal/core/auth/jwt"
	"backend/internal/core/cache"
	"backend/internal/core/config"
	"backend/internal/core/kafka"
	"backend/internal/core/logger"
	"backend/internal/core/repository/postgres/pool/pgx"
	"backend/internal/core/transport/grpc/interceptors"
	finance_repo "backend/internal/features/finance/repository/postgres"
	finance_service "backend/internal/features/finance/service"
	finance_grpc "backend/internal/features/finance/transport/grpc"
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

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

	logger.Debug("initializing finance service")
	financeRepository := finance_repo.NewFinanceRepository(pool)
	financeService := finance_service.NewFinanceService(financeRepository, pool, redisClient, kafkaProducer)

	logger.Debug("initializing finance grpc server")
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.RequestIDInterceptor(),
			interceptors.LoggerInterceptor(logger),
			interceptors.AuthInterceptor(jwtManager),
			interceptors.TraceInterceptor(logger.Logger)),
	)

	financeServer := finance_grpc.NewFinanceServer(financeService, logger)
	finance_grpc.RegisterFinanceServiceServer(grpcServer, financeServer)

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
	logger.Warn("shutting down gRPC server", zap.String("address", ":50053"))
	grpcServer.GracefulStop()
	logger.Warn("gRPC server stopped", zap.String("address", ":50053"))

	//rest api
	//financeHandler := finance_http.NewFinanceHandler(financeService, jwtManager)
	//
	//logger.Debug("initializing http server")
	//httpServer := server.NewHTTPServer(
	//	server.NewConfigMust(),
	//	logger,
	//	core_http_middleware.CORS(),
	//	core_http_middleware.RequestID(),
	//	core_http_middleware.Logger(logger),
	//	core_http_middleware.Trace(),
	//	core_http_middleware.Panic(),
	//)
	//
	//httpServer.RegisterRoutes(financeHandler.Routes()...)
	//
	//if err := httpServer.Run(ctx); err != nil {
	//	logger.Error("Failed to start server", zap.Error(err))
	//}
}
