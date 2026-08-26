package service_admin

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc/metadata"

	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/pkg/pagination"
	userpb "github.com/daniiiiiiiiiiil/finance-microservices/admin-service/proto/users/gen"
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

	var phoneNumber *string
	if resp.PhoneNumber != "" {
		phoneNumber = &resp.PhoneNumber
	}

	user = domain.User{
		ID:          int(resp.Id),
		Version:     int(resp.Version),
		FullName:    resp.FullName,
		Email:       resp.Email,
		PhoneNumber: phoneNumber,
		IsAdmin:     resp.IsAdmin,
	}

	if err := s.redis.Set(ctx, key, user, 10*time.Minute); err != nil {
		log.Printf("redis set error: %v", err)
	}

	return user, nil
}

func (s *AdminService) GetUsers(ctx context.Context, limit, offset int) ([]domain.User, error) {
	limit, offset = pagination.LimitOffset(limit, offset)

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		auth := md.Get("authorization")
		if len(auth) > 0 {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", auth[0])
		}
	}

	grpcReq := &userpb.ListUsersRequest{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	resp, err := s.userClient.ListUsers(ctx, grpcReq)
	if err != nil {
		return nil, fmt.Errorf("get users: %w", err)
	}
	domainUsers := make([]domain.User, len(resp.Users))
	for i, user := range resp.Users {
		var phoneNumber *string
		if user.PhoneNumber != nil {
			phoneNumber = user.PhoneNumber
		}
		domainUsers[i] = domain.User{
			ID:          user.ID,
			FullName:    user.FullName,
			Email:       user.Email,
			PhoneNumber: phoneNumber,
			IsAdmin:     user.IsAdmin,
		}
	}
	return domainUsers, nil
}
