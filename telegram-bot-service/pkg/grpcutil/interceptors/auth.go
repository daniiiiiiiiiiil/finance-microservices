package interceptors

import (
	"context"
	"strings"

	"github.com/daniiiiiiiiiiil/finance-microservices/telegram-bot-service/internal/core/auth/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
	ClaimsKey contextKey = "claims"
)

func AuthInterceptor(jwtManager *jwt.Manager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing authorization metadata")
		}

		var tokenString string
		if auth := md.Get("authorization"); len(auth) > 0 {
			parts := strings.Split(auth[0], " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}
		if tokenString == "" {
			if cookie := md.Get("cookie"); len(cookie) > 0 {
				for _, part := range strings.Split(cookie[0], "; ") {
					if strings.HasPrefix(part, "token=") {
						tokenString = strings.TrimPrefix(part, "token=")
						break
					}
				}
			}
		}
		if tokenString == "" {
			return nil, status.Error(codes.Unauthenticated, "missing authorization token")
		}

		claims, err := jwtManager.Validate(tokenString)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, ClaimsKey, claims)
		return handler(ctx, req)
	}
}

func GetUserID(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(UserIDKey).(int)
	return userID, ok
}

func GetClaims(ctx context.Context) (*jwt.Claims, bool) {
	claims, ok := ctx.Value(ClaimsKey).(*jwt.Claims)
	return claims, ok
}
