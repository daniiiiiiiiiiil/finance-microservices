package core_http_middleware

import (
	"backend/internal/core/logger"
	"backend/internal/core/transport/http/response"
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const requestIDHeader = "X-Request-Id"

func CORS() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", "POST, GET, PATCH, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(requestIDHeader)
			if requestID == "" {
				requestID = uuid.NewString()
			}
			w.Header().Set(requestIDHeader, requestID)
			r.Header.Set(requestIDHeader, requestID)

			next.ServeHTTP(w, r)
		})
	}
}

func Logger(log *logger.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(requestIDHeader)

			l := log.With(
				zap.String("request_id", requestID),
				zap.String("url", r.URL.String()),
			)
			ctx := logger.ToContext(r.Context(), l)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func Trace() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			log, ok := ctx.Value(logger.LoggerContextKey).(*logger.Logger)
			if !ok {
				log, _ = logger.NewLogger(logger.NewConfigMust())
				ctx = context.WithValue(ctx, logger.LoggerContextKey, log)
				r = r.WithContext(ctx)
			}

			rw := response_core.NewResponseWriter(w)
			before := time.Now()
			log.Debug("incoming http request", zap.String("url", r.URL.String()), zap.String("method", r.Method))
			next.ServeHTTP(rw, r)
			log.Debug("done http request", zap.Int("status_code", rw.GetStatusCode()), zap.Duration("elapsed", time.Since(before)))
		})
	}
}

func Panic() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			log, ok := ctx.Value(logger.LoggerContextKey).(*logger.Logger)
			if !ok {
				log, _ = logger.NewLogger(logger.NewConfigMust())
				ctx = context.WithValue(ctx, logger.LoggerContextKey, log)
				r = r.WithContext(ctx)
			}

			responseHandler := response_core.NewHTTPResponseHandler(log, w)
			defer func() {
				if p := recover(); p != nil {
					responseHandler.PanicResponse(p, "unexpected panic")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
