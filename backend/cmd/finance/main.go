package main

import (
	"backend/internal/core/auth/jwt"
	"backend/internal/core/config"
	"backend/internal/core/logger"
	"backend/internal/core/repository/postgres/pool/pgx"
	core_http_middleware "backend/internal/core/transport/http/middleware"
	"backend/internal/core/transport/http/server"
	admin_repo "backend/internal/features/admin/repository/postgres"
	admin_service "backend/internal/features/admin/service"
	admin_http "backend/internal/features/admin/transport/http"
	postgres_auth "backend/internal/features/auth/repository/postgres"
	service_auth "backend/internal/features/auth/service"
	http_auth "backend/internal/features/auth/transport/http"
	"backend/internal/features/auth/transport/http/dto"
	finance_repo "backend/internal/features/finance/repository/postgres"
	finance_service "backend/internal/features/finance/service"
	finance_http "backend/internal/features/finance/transport/http"
	"backend/internal/features/users/repository/postgres"
	"backend/internal/features/users/service"
	user_transport_http "backend/internal/features/users/transport/http"
	file_system_web "backend/internal/features/web/repository/file_system"
	service_web "backend/internal/features/web/service"
	http_web "backend/internal/features/web/transport/http"
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	_ "backend/docs"
)

// @title Golang Finance-microservices
// @version 1.0
// @description Finance REST-API scheme
// @host 127.0.0.1:5050
// @BasePath /api/v1
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

	jwtManager := jwt.NewJWTManager(cfg.JWTSecret, cfg.JWTDuration)
	logger.Debug("initializing feature", zap.String("feature", "auth"))
	authRepo := postgres_auth.NewAuthRepository(pool)
	authService := service_auth.NewAuthService(authRepo, jwtManager)
	createFirstAdmin(ctx, authService, logger)
	authTransportHTTP := http_auth.NewAuthHandler(authService)

	logger.Debug("initializing auth")

	logger.Debug("initializing feature", zap.String("feature", "users"))
	usersRepository := postgres.NewUserRepository(pool)
	usersService := service.NewUsersService(usersRepository, pool)
	usersTransportHTTP := user_transport_http.NewUsersHTTPHandler(usersService, jwtManager)

	logger.Debug("initializing feature", zap.String("feature", "finance"))
	financeRepository := finance_repo.NewFinanceRepository(pool)
	financeService := finance_service.NewFinanceService(financeRepository, pool)
	financeHandler := finance_http.NewFinanceHandler(financeService, jwtManager)

	logger.Debug("initializing feature", zap.String("feature", "admin"))
	adminRepository := admin_repo.NewAdminRepository(pool)
	adminService := admin_service.NewAdminService(adminRepository, pool)
	adminHandler := admin_http.NewAdminHandler(adminService, jwtManager)

	logger.Debug("initializing features", zap.String("features", "web"))
	webRepository := file_system_web.NewWebRepository()
	webService := service_web.NewWebService(webRepository)
	webTransport := http_web.NewWebHTTPHandler(webService)

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

	apiVersionRouter := server.NewAPIVersionRouter(server.ApiVersion1)
	apiVersionRouter.RegisterRoutes(authTransportHTTP.Routes()...)
	apiVersionRouter.RegisterRoutes(usersTransportHTTP.Routes()...)
	apiVersionRouter.RegisterRoutes(financeHandler.Routes()...)
	apiVersionRouter.RegisterRoutes(adminHandler.Routes()...)

	httpServer.RegisterRouters(apiVersionRouter) //apiVersionRouter2

	httpServer.RegisterRoutes(webTransport.Routes()...)

	httpServer.RegisterSwagger()

	if err := httpServer.Run(ctx); err != nil {
		logger.Error("Failed to start server", zap.Error(err))
	}
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
