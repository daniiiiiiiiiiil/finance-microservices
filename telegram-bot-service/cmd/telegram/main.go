package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/config"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/auth/jwt"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/cache"
	gatewayclient "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/clients/gateway"
	grpcserver "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/grpc"
	corekafka "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/kafka"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/repository/postgres/pool/pgx"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/telemetry"
	postgresrepo "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/features/repository/postgres"
	redisrepo "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/features/repository/redis"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/features/service"
	grpctransport "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/features/transport/grpc"
	httptransport "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/features/transport/http"
	kafkatransport "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/features/transport/kafka"
	telegramtransport "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/features/transport/telegram"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/features/webapp"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/grpcutil/interceptors"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

const (
	serviceName          = "telegram-bot-service"
	updateWorkers        = 4
	updateQueueSize      = 256
	updateProcessTimeout = 3 * time.Minute
	shutdownTimeout      = 30 * time.Second
)

func main() {
	cfg := config.NewConfigMust()
	time.Local = cfg.TimeZone

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger, err := logger.NewLogger(logger.NewConfigMust())
	if err != nil {
		os.Exit(1)
	}
	defer logger.Close()

	shutdownTracer, err := telemetry.InitTracer(serviceName, "jaeger:4317")
	if err != nil {
		logger.Fatal("telemetry init error", zap.Error(err))
	}
	defer shutdownTracer()

	logger.Debug("initializing postgres connection pool")
	connectionPool, err := pgx.NewPool(ctx, pgx.NewConfigMust())
	if err != nil {
		logger.Fatal("postgres connection error", zap.Error(err))
	}
	defer connectionPool.Close()

	logger.Debug("initializing redis")
	redisClient := cache.NewRedisClient(cfg.RedisAddr)
	defer redisClient.Close()

	repository := postgresrepo.NewRepository(connectionPool)
	dialogStore := redisrepo.NewDialogStore(redisClient, cfg.DialogTTL)
	dedupStore := redisrepo.NewDedupStore(redisClient)

	accountService := service.NewAccountService(repository, repository, logger, cfg.LinkCodeTTL)
	dialogService := service.NewDialogService(dialogStore)

	logger.Debug("initializing telegram api")
	telegramAPI, err := telegramtransport.NewAdapter(cfg.TelegramToken, logger)
	if err != nil {
		logger.Fatal("telegram api init error", zap.Error(err))
	}

	gatewayTokens := jwt.NewManager(cfg.JWTSecret, cfg.ServiceTokenTTL)
	webAppTokens := jwt.NewManager(cfg.JWTSecret, cfg.WebAppTokenTTL)
	grpcTokens := jwt.NewManager(cfg.JWTSecret, cfg.JWTDuration)

	gatewayClient := gatewayclient.NewClient(cfg.GatewayBaseURL, gatewayTokens, cfg.HTTPTimeout, logger)

	notificationService := service.NewNotificationService(repository, repository, telegramAPI, logger, cfg.NotificationTries)

	router := telegramtransport.NewRouter(
		telegramAPI,
		accountService,
		dialogService,
		gatewayClient,
		logger,
		cfg.WebAppURL(),
	)
	runner := telegramtransport.NewRunner(
		router,
		dedupStore,
		logger,
		updateWorkers,
		updateQueueSize,
		cfg.UpdateDedupTTL,
		updateProcessTimeout,
	)
	runner.Start(ctx)
	httpServer := httptransport.NewServer(
		cfg,
		runner,
		accountService,
		webapp.NewVerifier(cfg.TelegramToken, cfg.TelegramWebAppMaxAge),
		webAppTokens,
		logger,
		connectionPool,
		redisClient,
	)

	go func() {
		if err := httpServer.Start(); err != nil {
			logger.Error("telegram http server error", zap.Error(err))
		}
	}()

	metricsServer := &http.Server{
		Addr:    cfg.TelegramMetricsAddr,
		Handler: promhttp.Handler(),
	}
	go func() {
		logger.Info("starting metrics server", zap.String("addr", cfg.TelegramMetricsAddr))
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("metrics server error", zap.Error(err))
		}
	}()

	logger.Debug("initializing kafka consumer")
	kafkaConsumer := corekafka.NewConsumer(corekafka.NewConfig(), logger)
	kafkaHandlers := kafkatransport.NewHandlers(notificationService, dedupStore, logger, cfg.EventDedupTTL)
	if err := kafkaHandlers.Register(kafkaConsumer); err != nil {
		logger.Fatal("register kafka handlers", zap.Error(err))
	}
	go func() {
		if err := kafkaConsumer.Start(ctx); err != nil {
			logger.Error("kafka consumer error", zap.Error(err))
		}
	}()
	defer func() {
		if err := kafkaConsumer.Close(); err != nil {
			logger.Warn("close kafka consumer", zap.Error(err))
		}
	}()

	logger.Debug("initializing grpc server")
	grpcServer := grpcserver.NewGRPCServer(
		cfg,
		interceptors.RequestIDInterceptor(),
		interceptors.LoggerInterceptor(logger),
		interceptors.AuthInterceptor(grpcTokens),
		interceptors.MetricsInterceptor(serviceName),
		interceptors.TraceInterceptor(),
	)
	grpcHandlers := grpctransport.NewServer(accountService, logger)
	grpctransport.RegisterTelegramServer(grpcServer, grpcHandlers)

	grpcListener, err := net.Listen("tcp", cfg.TelegramGRPCAddr)
	if err != nil {
		logger.Fatal("failed to listen grpc", zap.Error(err))
	}
	go func() {
		logger.Info("starting grpc server", zap.String("addr", cfg.TelegramGRPCAddr))
		if err := grpcServer.Serve(grpcListener); err != nil {
			logger.Error("grpc server error", zap.Error(err))
		}
	}()

	switch cfg.TelegramMode {
	case config.ModePolling:
		go func() {
			logger.Info("starting telegram long polling")
			if err := telegramAPI.StartPolling(ctx, runner.SubmitRaw); err != nil {
				logger.Error("telegram polling error", zap.Error(err))
			}
		}()
	case config.ModeWebhook:
		go registerWebhook(ctx, telegramAPI, cfg, logger)
	}

	if identity, err := telegramIdentity(ctx, telegramAPI); err == nil {
		logger.Info("telegram bot authorized", zap.String("username", identity.Username))
	} else {
		logger.Warn("telegram getMe failed", zap.Error(err))
	}

	<-ctx.Done()
	logger.Warn("shutting down telegram bot service")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("http server shutdown error", zap.Error(err))
	}
	if err := metricsServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("metrics server shutdown error", zap.Error(err))
	}

	grpcDone := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(grpcDone)
	}()
	select {
	case <-grpcDone:
		logger.Info("grpc server gracefully stopped")
	case <-shutdownCtx.Done():
		logger.Warn("grpc server shutdown timed out")
		grpcServer.Stop()
	}

	runner.Stop(shutdownCtx)
	logger.Info("shutdown complete")
}

func registerWebhook(ctx context.Context, telegramAPI *telegramtransport.Adapter, cfg *config.Config, logger *logger.Logger) {
	delay := 2 * time.Second
	for attempt := 1; attempt <= 5; attempt++ {
		if err := telegramAPI.SetWebhook(ctx, cfg.TelegramWebhookURL, cfg.TelegramWebhookSecret); err == nil {
			logger.Info("telegram webhook configured", zap.String("url", cfg.TelegramWebhookURL))
			return
		} else if ctx.Err() != nil {
			return
		} else {
			logger.Warn("telegram webhook setup failed", zap.Int("attempt", attempt), zap.Error(err))
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}
		delay *= 2
	}
	logger.Error("telegram webhook setup exhausted retries")
}

func telegramIdentity(ctx context.Context, telegramAPI *telegramtransport.Adapter) (ports.BotIdentity, error) {
	identityCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return telegramAPI.GetMe(identityCtx)
}
