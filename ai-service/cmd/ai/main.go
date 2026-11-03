package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/core/repository/postgres/pool/pgx"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/config"
	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/core/auth/jwt"
	corekafka "github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/core/kafka"
	postgresrepo "github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/features/repository/postgres"
	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/features/service"
	shoppinggrpc "github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/features/transport/grpc"
	transportkafka "github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/features/transport/kafka"
	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/pkg/grpcutil/interceptors"
	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/pkg/logger"
	aipb "github.com/daniiiiiiiiiiil/finance-microservices/ai-service/proto/gen/v1"
)

func main() {
	cfg := config.NewConfigMust()
	if err := cfg.Validate(); err != nil {
		panic(fmt.Sprintf("invalid config: %v", err))
	}
	time.Local = cfg.TimeZone

	log := logger.NewMust()
	defer log.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	jwtManager := jwt.NewJWTManager(cfg.JWTSecret, cfg.JWTDuration)

	llm := service.NewGroqClient(
		cfg.GroqAPIKey, cfg.GroqBaseURL, cfg.GroqModel,
		cfg.GroqTimeout, cfg.GroqMaxRetries, log,
	)
	tavily := service.NewTavilyClient(
		cfg.TavilyAPIKey, cfg.TavilyBaseURL, cfg.TavilyTimeout, log,
	)
	geo := service.NewOSMClient(
		cfg.NominatimURL, cfg.OSRMURL, cfg.NominatimUserAgent,
		cfg.NominatimTimeout, log,
	)

	var planRepo *postgresrepo.PlanRepository
	if cfg.PostgresHost != "" {
		pool, err := pgx.NewPool(ctx, pgx.Config{
			Host:     cfg.PostgresHost,
			Port:     cfg.PostgresPort,
			User:     cfg.PostgresUser,
			Password: cfg.PostgresPassword,
			Database: cfg.PostgresDatabase,
			Timeout:  cfg.PostgresTimeout,
		})
		if err != nil {
			log.Fatal("postgres connect error", zap.Error(err))
		}
		defer pool.Close()
		planRepo = postgresrepo.NewPlanRepository(pool)
	}

	var eventPublisher *transportkafka.EventPublisher
	kafkaConfig := corekafka.NewConfig()
	kafkaProducer := corekafka.NewProducer(kafkaConfig, log)
	defer kafkaProducer.Close()
	eventPublisher = transportkafka.NewEventPublisher(kafkaProducer)

	svc := service.NewService(
		llm,
		tavily,
		geo,
		geo,
		planRepo,
		eventPublisher,
		service.CostConfig{
			FuelConsumptionLPer100KM: cfg.FuelConsumptionLPer100KM,
			FuelPricePerLiter:        cfg.FuelPricePerLiter,
			TaxiPricePerKM:           cfg.TaxiPricePerKM,
			Mode:                     cfg.TravelCostMode,
		},
		log,
	)

	kafkaConsumer := corekafka.NewConsumer(kafkaConfig, log)
	consumer := transportkafka.NewConsumer(kafkaConsumer)
	handlers := transportkafka.NewHandlers(svc, log)
	if err := handlers.Register(consumer); err != nil {
		log.Fatal("register kafka handlers", zap.Error(err))
	}
	go func() {
		if err := consumer.Start(ctx); err != nil {
			log.Error("kafka consumer error", zap.Error(err))
		}
	}()
	defer consumer.Close()

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.RequestIDInterceptor(),
			interceptors.LoggerInterceptor(log),
			interceptors.AuthInterceptor(jwtManager),
			interceptors.TraceInterceptor(),
		),
	)
	aiServer := shoppinggrpc.NewServer(svc, log)
	aipb.RegisterAIServiceServer(grpcServer, aiServer)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}
	log.Info("starting ai gRPC server", zap.String("port", cfg.GRPCPort))

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("grpc server error", zap.Error(err))
		}
	}()

	metricsServer := &http.Server{
		Addr:    ":" + cfg.MetricsPort,
		Handler: promhttp.Handler(),
	}
	go func() {
		log.Info("starting metrics server", zap.String("port", cfg.MetricsPort))
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("metrics server error", zap.Error(err))
		}
	}()

	<-ctx.Done()
	log.Info("shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	done := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(done)
	}()
	select {
	case <-done:
		log.Info("gRPC server stopped gracefully")
	case <-shutdownCtx.Done():
		log.Warn("gRPC server stop timeout, forcing stop")
		grpcServer.Stop()
	}

	if err := metricsServer.Shutdown(shutdownCtx); err != nil {
		log.Error("metrics server shutdown error", zap.Error(err))
	}
	log.Info("shutdown complete")
}
