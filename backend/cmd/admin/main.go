package main

import (
	"backend/internal/core/auth/jwt"
	"backend/internal/core/cache"
	financeСlient "backend/internal/core/clients/finance"
	usersclient "backend/internal/core/clients/users"
	"backend/internal/core/config"
	"backend/internal/core/kafka"
	"backend/internal/core/logger"
	"backend/internal/core/repository/postgres/pool/pgx"
	"backend/internal/core/transport/grpc/interceptors"
	admin_repo "backend/internal/features/admin/repository/postgres"
	admin_service "backend/internal/features/admin/service"
	admingrpc "backend/internal/features/admin/transport/grpc"
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

	logger.Debug("Initializing user client")
	usersClient, err := usersclient.NewUsersClient("users:50052")
	if err != nil {
		logger.Fatal("failed to initialize user client", zap.Error(err))
	}
	defer usersClient.Close()

	financeClient, err := financeСlient.NewFinanceClient("finance:50053")
	if err != nil {
		logger.Fatal("failed to connect to finance service", zap.Error(err))
	}
	defer financeClient.Close()

	logger.Debug("initializing admin service")
	adminRepository := admin_repo.NewAdminRepository(pool)
	adminService := admin_service.NewAdminService(adminRepository, pool, redisClient, kafkaProducer, usersClient, financeClient)
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.RequestIDInterceptor(),
			interceptors.LoggerInterceptor(logger),
			interceptors.AuthInterceptor(jwtManager),
			interceptors.TraceInterceptor(logger.Logger),
		))
	adminServer := admingrpc.NewAdminServer(adminService, logger)
	admingrpc.RegisterAdminServer(grpcServer, adminServer)

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
	logger.Warn("shutting down gRPC server", zap.String("address", ":50054"))
	grpcServer.GracefulStop()
	logger.Info("gRPC server stopped", zap.String("address", ":50054"))

	//rest api
	//adminHandler := admin_http.NewAdminHandler(adminService, jwtManager)
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
	//httpServer.RegisterRoutes(adminHandler.Routes()...)
	//
	//if err := httpServer.Run(ctx); err != nil {
	//	logger.Error("Failed to start server", zap.Error(err))
	//}
}
