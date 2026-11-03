package interceptors

import (
	"context"
	"strings"

	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/core/auth/jwt"
	"github.com/daniiiiiiiiiiil/finance-microservices/ai-service/internal/core/ports"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func AuthInterceptor(jwtManager *jwt.JWTManager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var tokenString string

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		auth := md.Get("authorization")
		if len(auth) > 0 {
			fields := strings.Fields(auth[0])
			if len(fields) == 2 && fields[0] == "Bearer" {
				tokenString = fields[1]
			}
		}

		if tokenString == "" {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		claims, err := jwtManager.Validate(tokenString)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		ctx = ports.WithUser(ctx, ports.UserContext{
			UserID:  claims.UserID,
			Email:   claims.Email,
			IsAdmin: claims.IsAdmin,
		})

		return handler(ctx, req)
	}
}

func GetUserID(ctx context.Context) (int, bool) {
	u, ok := ports.UserFromContext(ctx)
	return u.UserID, ok
}
