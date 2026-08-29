package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"

	"github.com/daniiiiiiiiiiil/finance-microservices/gateway-service/config"
	"github.com/daniiiiiiiiiiil/finance-microservices/gateway-service/internal/core/auth/jwt"
	grpcclient "github.com/daniiiiiiiiiiil/finance-microservices/gateway-service/internal/core/grpc"
	"github.com/daniiiiiiiiiiil/finance-microservices/gateway-service/pkg/logger"
	http2 "github.com/daniiiiiiiiiiil/finance-microservices/gateway-service/pkg/middleware/http"
	adminpb "github.com/daniiiiiiiiiiil/finance-microservices/gateway-service/proto/admin/gen"
	authpb "github.com/daniiiiiiiiiiil/finance-microservices/gateway-service/proto/auth/gen"
	currencypb "github.com/daniiiiiiiiiiil/finance-microservices/gateway-service/proto/currency/gen"
	financepb "github.com/daniiiiiiiiiiil/finance-microservices/gateway-service/proto/finance/gen"
	userpb "github.com/daniiiiiiiiiiil/finance-microservices/gateway-service/proto/users/gen"
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

	logger.Debug("time zone", zap.Any("timezone", time.Local))

	jwtManager := jwt.NewJWTManager(cfg.JWTSecret, cfg.JWTDuration)
	logger.Debug("JWT Manager initialized for Gateway")

	authConn, err := grpcclient.NewGRPCClient("auth:50051", cfg)
	if err != nil {
		logger.Fatal("fail to dial auth", zap.Error(err))
	}
	defer authConn.Close()

	userConn, err := grpcclient.NewGRPCClient("users:50052", cfg)
	if err != nil {
		logger.Fatal("fail to dial users", zap.Error(err))
	}
	defer userConn.Close()

	financeConn, err := grpcclient.NewGRPCClient("finance:50053", cfg)
	if err != nil {
		logger.Fatal("fail to dial finance", zap.Error(err))
	}
	defer financeConn.Close()

	adminConn, err := grpcclient.NewGRPCClient("admin:50054", cfg)
	if err != nil {
		logger.Fatal("fail to dial admin", zap.Error(err))
	}
	defer adminConn.Close()

	currencyConn, err := grpcclient.NewGRPCClient("currency:50055", cfg)
	if err != nil {
		logger.Fatal("fail to dial currency", zap.Error(err))
	}
	defer currencyConn.Close()

	mux := runtime.NewServeMux(
		runtime.WithMetadata(func(ctx context.Context, r *http.Request) metadata.MD {
			md := metadata.MD{}
			if auth := r.Header.Get("Authorization"); auth != "" {
				md.Set("authorization", auth)
			}
			return md
		}))

	if err := authpb.RegisterAuthServiceHandler(ctx, mux, authConn); err != nil {
		logger.Fatal("fail to register auth service", zap.Error(err))
	}

	if err := financepb.RegisterFinanceServiceHandler(ctx, mux, financeConn); err != nil {
		logger.Fatal("fail to register finance service", zap.Error(err))
	}

	if err := userpb.RegisterUserServiceHandler(ctx, mux, userConn); err != nil {
		logger.Fatal("fail to register user service", zap.Error(err))
	}

	if err := adminpb.RegisterAdminServiceHandler(ctx, mux, adminConn); err != nil {
		logger.Fatal("fail to register admin service", zap.Error(err))
	}

	if err := currencypb.RegisterCurrencyServiceHandler(ctx, mux, currencyConn); err != nil {
		logger.Fatal("fail to register currency service", zap.Error(err))
	}

	otelMux := otelhttp.NewHandler(mux, "gateway-http")

	httpMux := http.NewServeMux()

	httpMux.Handle("/metrics", promhttp.Handler())

	httpMux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/swagger.json"),
		httpSwagger.UIConfig(map[string]string{
			"requestInterceptor": `(req) => {
                const token = localStorage.getItem('swagger_token');
                if (token) {
                    req.headers['Authorization'] = 'Bearer ' + token;
                }
                req.credentials = 'include';
                return req;
            }`,
		}),
	))

	httpMux.HandleFunc("/swagger/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, "./docs/swagger.json")
	})

	httpMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("ok"))
		if err != nil {
			logger.Error("failed to write response", zap.Error(err))
			return
		}
	})

	httpMux.Handle("/api/", otelMux)

	handler := http2.ChainMiddlewares(
		httpMux,
		http2.CORS(),
		http2.Auth(jwtManager),
		http2.RequestID(),
		http2.Logger(logger),
		http2.Trace(),
		http2.Panic(),
	)

	server := &http.Server{
		Addr:    ":8081",
		Handler: handler,
	}

	go func() {
		logger.Warn("http server start", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("fail to start http server", zap.Error(err))
		}
	}()

	<-ctx.Done()

	logger.Warn("shutting down gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	done := make(chan struct{})
	go func() {
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("failed to shutdown http server", zap.Error(err))
		}
		close(done)
	}()

	select {
	case <-done:
		logger.Info("HTTP server stopped gracefully")
	case <-shutdownCtx.Done():
		logger.Warn("HTTP server stop timeout, forcing stop")
		if err := server.Close(); err != nil {
			logger.Error("failed to force close http server", zap.Error(err))
		}
	}

	logger.Info("shutdown complete")
}
