package main

import (
	"backend/internal/core/auth/jwt"
	"backend/internal/core/config"
	"backend/internal/core/logger"
	core_http_middleware "backend/internal/core/transport/http/middleware"
	"backend/internal/core/transport/http/server"
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"

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

	jwtManager := jwt.NewJWTManager(cfg.JWTSecret, cfg.JWTDuration)
	logger.Debug("JWT Manager initialized for Gateway")

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

	registerProxy(httpServer, "/api/v1/auth", "http://localhost:5051", jwtManager, false)   // Auth — без проверки (логин/регистрация)
	registerProxy(httpServer, "/api/v1/users", "http://localhost:5052", jwtManager, true)   // Users — с проверкой
	registerProxy(httpServer, "/api/v1/finance", "http://localhost:5050", jwtManager, true) // Finance — с проверкой
	registerProxy(httpServer, "/api/v1/admin", "http://localhost:5053", jwtManager, true)   // Admin — с проверкой

	httpServer.RegisterSwagger()

	httpServer.RegisterRoutes(server.Route{
		Method: http.MethodGet,
		Path:   "/healthz",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		},
	})

	if err := httpServer.Run(ctx); err != nil {
		logger.Error("Failed to start server", zap.Error(err))
	}
}

func registerProxy(
	httpServer *server.HTTPServer,
	path string,
	target string,
	jwtManager *jwt.JWTManager,
	requireAuth bool,
) {
	remote, err := url.Parse(target)
	if err != nil {
		panic(err)
	}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			originalPath := req.URL.Path

			req.URL.Path = strings.TrimPrefix(req.URL.Path, "/api/v1")
			if req.URL.Path == "" {
				req.URL.Path = "/"
			}

			req.URL.Scheme = remote.Scheme
			req.URL.Host = remote.Host

			fmt.Printf("proxying: original=%s -> target=%s%s\n",
				originalPath, remote.String(), req.URL.Path)
		},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		if !requireAuth {
			proxy.ServeHTTP(w, r)
			return
		}

		tokenString := extractToken(r)
		if tokenString == "" {
			http.Error(w, "missing authorization", http.StatusUnauthorized)
			return
		}

		claims, err := jwtManager.Validate(tokenString)
		if err != nil {
			fmt.Printf("Gateway: invalid token: %v\n", err)
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		fmt.Printf("Gateway: token valid for user %d\n", claims.UserID)

		r.Header.Set("X-User-Id", strconv.Itoa(claims.UserID))
		r.Header.Set("X-Is-Admin", strconv.FormatBool(claims.IsAdmin))

		proxy.ServeHTTP(w, r)
	}

	methods := []string{
		http.MethodGet, http.MethodPost, http.MethodPut,
		http.MethodPatch, http.MethodDelete, http.MethodOptions,
	}

	for _, method := range methods {
		httpServer.RegisterRoutes(server.Route{
			Method: method,
			Path:   path + "/",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				handler(w, r)
			},
		})
	}
}

func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	cookie, err := r.Cookie("token")
	if err == nil {
		return cookie.Value
	}

	return ""
}
