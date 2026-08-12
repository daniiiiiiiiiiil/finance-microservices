package main

import (
	"backend/internal/core/auth/jwt"
	"backend/internal/core/cache"
	"backend/internal/core/config"
	"backend/internal/core/logger"
	"backend/internal/core/repository/postgres/pool/pgx"
	"backend/internal/core/transport/grpc/interceptors"
	postgres_auth "backend/internal/features/auth/repository/postgres"
	service_auth "backend/internal/features/auth/service"
	authgrpc "backend/internal/features/auth/transport/grpc"
	"backend/internal/features/auth/transport/http/dto"
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

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

	jwtManager := jwt.NewJWTManager(cfg.JWTSecret, cfg.JWTDuration)

	logger.Debug("initializing auth service")
	authRepo := postgres_auth.NewAuthRepository(pool)
	authService := service_auth.NewAuthService(authRepo, jwtManager, redisClient)
	createFirstAdmin(ctx, authService, logger)

	logger.Debug("initializing auth service gRPC")
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.RequestIDInterceptor(),
			interceptors.LoggerInterceptor(logger),
			interceptors.TraceInterceptor(logger.Logger)),
	)

	authServer := authgrpc.NewAuthServer(authService, logger)
	authgrpc.RegisterAuthServiceServer(grpcServer, authServer)

	reflection.Register(grpcServer)

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
	logger.Warn("shutting down gRPC server", zap.String("address", ":50051"))
	grpcServer.GracefulStop()
	logger.Warn("gRPC server stopped", zap.String("address", ":50051"))

	//rest api
	//authHandler := http_auth.NewAuthHandler(authService)
	//
	//logger.Debug("initializing http server")
	//
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
	//httpServer.RegisterRoutes(authHandler.Routes()...)
	//httpServer.RegisterSwagger()
	//
	//if err := httpServer.Run(ctx); err != nil {
	//	logger.Error("Failed to start server", zap.Error(err))
	//}
}

func createFirstAdmin(ctx context.Context, authService *service_auth.AuthService, log *logger.Logger) {
	exists, err := authService.AdminExists(ctx)
	if err != nil {
		log.Error("failed to check admin existence", zap.Error(err))
		return
	}

	if !exists {
		req := dto.RegisterRequest{
			FullName:    "Admin",
			Email:       "admin@finance.com",
			Password:    "admin123",
			PhoneNumber: "",
			IsAdmin:     true,
		}

		_, _, err := authService.Register(ctx, req)
		if err != nil {
			log.Error("failed to create first admin", zap.Error(err))
		} else {
			log.Info("first admin created successfully")
		}
	}
}
