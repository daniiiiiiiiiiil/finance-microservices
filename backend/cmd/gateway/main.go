package main

import (
	_ "backend/docs"
	"backend/internal/core/auth/jwt"
	"backend/internal/core/config"
	"backend/internal/core/logger"
	core_http_middleware "backend/internal/core/transport/http/middleware"
	userpb "backend/internal/features/users/transport/grpc/proto"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	adminpb "backend/internal/features/admin/transport/grpc"
	authpb "backend/internal/features/auth/transport/grpc"
	financepb "backend/internal/features/finance/transport/grpc"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
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

	authConn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("fail to dial", zap.Error(err))
	}
	defer authConn.Close()

	userConn, err := grpc.Dial("localhost:50052",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("fail to dial", zap.Error(err))
	}
	defer userConn.Close()

	financeConn, err := grpc.Dial("localhost:50053",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("fail to dial", zap.Error(err))
	}
	defer financeConn.Close()

	adminConn, err := grpc.Dial("localhost:50054",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("fail to dial", zap.Error(err))
	}
	defer adminConn.Close()

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

	httpMux := http.NewServeMux()

	httpMux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/swagger.json"),
		httpSwagger.UIConfig(map[string]string{
			"requestInterceptor": `(req) => {
                // Поддержка авторизации через Bearer токен
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
		w.Write([]byte("ok"))
	})

	httpMux.Handle("/api/", mux)

	handler := core_http_middleware.ChainMiddlewares(
		httpMux,
		core_http_middleware.CORS(),
		core_http_middleware.Auth(jwtManager),
		core_http_middleware.RequestID(),
		core_http_middleware.Logger(logger),
		core_http_middleware.Trace(),
		core_http_middleware.Panic(),
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
	logger.Warn("shutting down http server", zap.String("addr", server.Addr))
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("fail to shutdown http server", zap.Error(err))
	}
	logger.Info("http server stopped")
}
