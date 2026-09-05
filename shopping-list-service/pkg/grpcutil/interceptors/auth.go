package interceptors

import (
	"context"
	"strings"

	"github.com/daniiiiiiiiiiil/finance-microservices/shopping-list-service/internal/core/auth/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const (
	UserIDKey  contextKey = "user_id"
	ClaimsKey  contextKey = "claims"
	IsAdminKey contextKey = "is_admin"
)

func AuthInterceptor(jwtManager *jwt.JWTManager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var tokenString string

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header or cookie")
		}

		auth := md.Get("authorization")
		if len(auth) > 0 {
			parts := strings.Split(auth[0], " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		if tokenString == "" {
			cookie := md.Get("cookie")
			if len(cookie) > 0 {
				for _, c := range strings.Split(cookie[0], "; ") {
					if strings.HasPrefix(c, "token=") {
						tokenString = strings.TrimPrefix(c, "token=")
						break
					}
				}
			}
		}

		if tokenString == "" {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header or cookie")
		}

		claims, err := jwtManager.Validate(tokenString)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, ClaimsKey, claims)
		ctx = context.WithValue(ctx, IsAdminKey, claims.IsAdmin)

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

func IsAdmin(ctx context.Context) bool {
	claims, ok := GetClaims(ctx)
	if !ok {
		return false
	}
	return claims.IsAdmin
}
