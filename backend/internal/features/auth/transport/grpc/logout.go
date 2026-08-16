package grpc

import (
	"backend/internal/core/transport/grpc/interceptors"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (s *AuthServer) Logout(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	userID, ok := interceptors.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}
	s.logger.Debug("Logout", zap.Int("user_id", userID))

	md, _ := metadata.FromIncomingContext(ctx)
	auth := md.Get("authorization")
	if len(auth) > 0 {
		token := strings.TrimPrefix(auth[0], "Bearer ")
		if err := s.service.AddToBlacklist(ctx, token, 24*time.Hour); err != nil {
			return nil, status.Error(codes.Unauthenticated, "unauthenticated")
		}
	}
	return &empty.Empty{}, nil
}
