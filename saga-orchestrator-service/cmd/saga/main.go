package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/config"
	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/domain"
	corekafka "github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/kafka"
	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/repository/postgres/pool/pgx"
	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/saga"
	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/core/telemetry"
	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/features/orchestrator"
	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/features/repository/postgres"
	transportkafka "github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/internal/features/transport/kafka"
	"github.com/daniiiiiiiiiiil/finance-microservices/saga-orchestrator-service/pkg/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	cfg := config.NewConfigMust()
	time.Local = cfg.TimeZone

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	loggerInstance, err := logger.NewLogger(logger.NewConfigMust())
	if err != nil {
		log.Fatal("failed to init logger", err)
	}
	defer loggerInstance.Close()

	pool, err := pgx.NewPool(ctx, pgx.NewConfigMust())
	if err != nil {
		loggerInstance.Fatal("postgres connection error", zap.Error(err))
	}
	defer pool.Close()

	shutdown, err := telemetry.InitTracer("saga-orchestrator")
	if err != nil {
		loggerInstance.Fatal("telemetry error", zap.Error(err))
	}
	defer shutdown()

	kafkaConfig := corekafka.NewConfig()
	kafkaProducer := corekafka.NewProducer(kafkaConfig, loggerInstance)
	defer kafkaProducer.Close()

	eventPublisher := transportkafka.NewSagaEventPublisher(kafkaProducer, loggerInstance)

	sagaRepo := postgres.NewSagaRepository(pool)

	sagaRegistry := orchestrator.NewSagaRegistryImpl()

	sagaManager := saga.NewSagaManager(
		loggerInstance,
		sagaRepo,
		sagaRegistry,
		eventPublisher,
		pool,
	)
	defer func() {
		if err := sagaManager.Shutdown(ctx); err != nil {
			loggerInstance.Error("saga manager shutdown error", zap.Error(err))
		}
	}()

	loggerInstance.Debug("initializing outbox publisher")
	outboxPublisher := saga.NewOutboxPublisher(sagaRepo, eventPublisher, loggerInstance)
	outboxPublisher.Start(ctx)

	deleteUserSaga := orchestrator.NewDeleteUserSaga(loggerInstance, sagaManager, eventPublisher, sagaRepo)
	registerUserSaga := orchestrator.NewRegisterUserSaga(loggerInstance, sagaManager, eventPublisher, sagaRepo)

	if err := sagaRegistry.Register(domain.SagaTypeDeleteUser, deleteUserSaga.Definition()); err != nil {
		loggerInstance.Fatal("failed to register delete user saga", zap.Error(err))
	}
	if err := sagaRegistry.Register(domain.SagaTypeRegisterUser, registerUserSaga.Definition()); err != nil {
		loggerInstance.Fatal("failed to register register user saga", zap.Error(err))
	}

	kafkaConsumer := corekafka.NewConsumer(kafkaConfig, loggerInstance)

	sagaConsumer := transportkafka.NewSagaKafkaConsumer(kafkaConsumer, loggerInstance)
	defer sagaConsumer.Close()

	handlers := transportkafka.NewSagaHandlers(deleteUserSaga, registerUserSaga, loggerInstance)
	if err := handlers.RegisterHandlers(sagaConsumer); err != nil {
		loggerInstance.Fatal("failed to register handlers", zap.Error(err))
	}

	go func() {
		if err := sagaConsumer.Start(ctx); err != nil {
			loggerInstance.Error("kafka consumer error", zap.Error(err))
		}
	}()

	if err := sagaManager.ResumeActiveSagas(ctx); err != nil {
		loggerInstance.Error("failed to resume active sagas", zap.Error(err))
	}

	metricsServer := http.Server{Addr: ":9098", Handler: promhttp.Handler()}
	go func() {
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			loggerInstance.Error("metrics server error", zap.Error(err))
		}
	}()

	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := sagaConsumer.Close(); err != nil {
		loggerInstance.Error("kafka consumer close error", zap.Error(err))
	}

	if err := metricsServer.Shutdown(shutdownCtx); err != nil {
		loggerInstance.Error("metrics server shutdown error", zap.Error(err))
	}
}
