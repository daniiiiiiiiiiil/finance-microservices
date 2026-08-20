package service_admin

import (
	"backend/internal/core/domain"
	"backend/internal/core/pagination"
	userpb "backend/internal/features/users/transport/grpc/proto"
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"
)

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
