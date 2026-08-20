package core_http_middleware

import (
	"backend/internal/core/auth/jwt"
	"context"
	"net/http"
	"strings"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
	ClaimsKey contextKey = "claims"
)

func Auth(jwtManager *jwt.JWTManager) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/healthz" ||
				r.URL.Path == "/metrics" ||
				strings.HasPrefix(r.URL.Path, "/swagger/") ||
				strings.HasPrefix(r.URL.Path, "/api/v1/auth/register") ||
				strings.HasPrefix(r.URL.Path, "/api/v1/auth/login") ||
				strings.HasPrefix(r.URL.Path, "/api/v1/auth/logout") {
				next.ServeHTTP(w, r)
				return
			}

			var tokenString string

			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					tokenString = parts[1]
				}
			}

			if tokenString == "" {
				cookie, err := r.Cookie("token")
				if err == nil {
					tokenString = cookie.Value
				}
			}

			if tokenString == "" {
				http.Error(w, "missing authorization header or cookie", http.StatusUnauthorized)
				return
			}

			claims, err := jwtManager.Validate(tokenString)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, ClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})

	}
}

func GetUserID(r *http.Request) (int, bool) {
	userID, ok := r.Context().Value(UserIDKey).(int)
	return userID, ok
}

func GetClaims(r *http.Request) (*jwt.Claims, bool) {
	claims, ok := r.Context().Value(ClaimsKey).(*jwt.Claims)
	return claims, ok
}

func IsAdmin(r *http.Request) bool {
	claims, ok := GetClaims(r)
	if !ok {
		return false
	}
	return claims.IsAdmin
}
