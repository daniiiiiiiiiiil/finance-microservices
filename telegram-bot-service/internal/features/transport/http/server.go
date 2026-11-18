package http

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/config"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/auth/jwt"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/ports"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/features/service"
	telegramtransport "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/features/transport/telegram"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/features/webapp"
	webappassets "github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/features/webapp/assets"
	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/pkg/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

const (
	maxWebhookBodyBytes = 1 << 20
	maxSessionBodyBytes = 64 << 10
	webhookSecretHeader = "X-Telegram-Bot-Api-Secret-Token"
)

type HealthChecker interface {
	Ping(ctx context.Context) error
}

type UpdateSubmitter interface {
	SubmitRaw(ctx context.Context, raw []byte) error
}

type Server struct {
	config     *config.Config
	runner     UpdateSubmitter
	accounts   ports.AccountServiceInterface
	verifier   *webapp.Verifier
	tokens     *jwt.Manager
	logger     *logger.Logger
	pool       HealthChecker
	cache      HealthChecker
	httpServer *http.Server
}

func NewServer(
	cfg *config.Config,
	runner UpdateSubmitter,
	accounts ports.AccountServiceInterface,
	verifier *webapp.Verifier,
	tokens *jwt.Manager,
	logger *logger.Logger,
	pool HealthChecker,
	cache HealthChecker,
) *Server {
	server := &Server{
		config:   cfg,
		runner:   runner,
		accounts: accounts,
		verifier: verifier,
		tokens:   tokens,
		logger:   logger,
		pool:     pool,
		cache:    cache,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/telegram/webhook", server.handleWebhook)
	mux.HandleFunc(cfg.TelegramWebAppPath, server.handleWebAppPage)
	mux.Handle(cfg.TelegramWebAppPath+"/assets/", http.StripPrefix(cfg.TelegramWebAppPath+"/assets/", webappassets.Handler()))
	mux.HandleFunc(cfg.TelegramWebAppPath+"/session", server.handleWebAppSession)
	mux.HandleFunc("/healthz", server.handleHealth)
	mux.Handle("/metrics", promhttp.Handler())

	server.httpServer = &http.Server{
		Addr:              cfg.TelegramHTTPAddr,
		Handler:           server.withRequestLogging(mux),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	return server
}

func (s *Server) Start() error {
	s.logger.Info("telegram http server started", zap.String("addr", s.config.TelegramHTTPAddr))
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("start http server: %w", err)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	if s.config.TelegramWebhookSecret == "" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "webhook is disabled"})
		return
	}
	provided := r.Header.Get(webhookSecretHeader)
	if subtle.ConstantTimeCompare([]byte(provided), []byte(s.config.TelegramWebhookSecret)) != 1 {
		s.logger.Warn("telegram webhook rejected: invalid secret token")
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid secret token"})
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxWebhookBodyBytes))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cannot read body"})
		return
	}
	if len(body) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "empty body"})
		return
	}

	if err := s.runner.SubmitRaw(r.Context(), body); err != nil {
		if errors.Is(err, telegramtransport.ErrQueueFull) {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "busy"})
			return
		}
		s.logger.Warn("telegram webhook update rejected", zap.Error(err))
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid update"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

type sessionRequest struct {
	InitData string `json:"initData"`
}

type sessionResponse struct {
	Token      string     `json:"token"`
	ExpiresIn  int64      `json:"expiresIn"`
	GatewayURL string     `json:"gatewayUrl"`
	User       userPublic `json:"user"`
}

type userPublic struct {
	ID        int64  `json:"id"`
	FirstName string `json:"firstName"`
	Username  string `json:"username"`
}

func (s *Server) handleWebAppSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, maxSessionBodyBytes))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cannot read body"})
		return
	}
	var request sessionRequest
	if err := json.Unmarshal(body, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	initData, err := s.verifier.Verify(request.InitData)
	if err != nil {
		if errors.Is(err, webapp.ErrExpiredInitData) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "init data expired"})
			return
		}
		s.logger.Warn("telegram web app init data rejected")
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid init data"})
		return
	}

	account, err := s.accounts.Resolve(r.Context(), initData.User.ID)
	if err != nil {
		if errors.Is(err, service.ErrNotLinked) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "telegram account is not linked"})
			return
		}
		s.logger.Error("resolve web app account", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	token, err := s.tokens.Generate(account.UserID)
	if err != nil {
		s.logger.Error("issue web app token", zap.Error(err))
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, sessionResponse{
		Token:      token,
		ExpiresIn:  int64(s.config.WebAppTokenTTL.Seconds()),
		GatewayURL: s.config.GatewayPublicURL,
		User: userPublic{
			ID:        initData.User.ID,
			FirstName: initData.User.FirstName,
			Username:  initData.User.Username,
		},
	})
}

func (s *Server) handleWebAppPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	webappassets.ServeIndex(w, r)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	status := map[string]string{"status": "ok", "service": "telegram-bot-service"}
	code := http.StatusOK
	if s.pool != nil {
		if err := s.pool.Ping(ctx); err != nil {
			status["postgres"] = "unavailable"
			status["status"] = "degraded"
			code = http.StatusServiceUnavailable
		} else {
			status["postgres"] = "ok"
		}
	}
	if s.cache != nil {
		if err := s.cache.Ping(ctx); err != nil {
			status["redis"] = "unavailable"
			status["status"] = "degraded"
			code = http.StatusServiceUnavailable
		} else {
			status["redis"] = "ok"
		}
	}
	writeJSON(w, code, status)
}

func (s *Server) withRequestLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		s.logger.Debug("http request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Duration("elapsed", time.Since(start)),
		)
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		return
	}
}
