package service_admin

import (
	"backend/internal/core/domain"
	userpb "backend/internal/features/users/transport/grpc/proto"
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc/metadata"
)

func (s *AdminService) GetUser(ctx context.Context, id int) (domain.User, error) {
	key := fmt.Sprintf("user:%d", id)
	var user domain.User

	err := s.redis.Get(ctx, key, &user)
	if err == nil {
		return user, nil
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		auth := md.Get("authorization")
		if len(auth) > 0 {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", auth[0])
		}
	}

	resp, err := s.userClient.GetUser(ctx, &userpb.GetUserRequest{
		Id: int32(id),
	})
	if err != nil {
		return domain.User{}, fmt.Errorf("get user from users service: %w", err)
	}

	user = domain.User{
		ID:          int(resp.Id),
		Version:     int(resp.Version),
		FullName:    resp.FullName,
		Email:       resp.Email,
		PhoneNumber: resp.PhoneNumber,
		IsAdmin:     resp.IsAdmin,
	}

	if err := s.redis.Set(ctx, key, user, 10*time.Minute); err != nil {
		log.Printf("redis set error: %v", err)
	}

	return user, nil
}
