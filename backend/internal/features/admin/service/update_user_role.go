package service_admin

import (
	"backend/internal/core/domain"
	userpb "backend/internal/features/users/transport/grpc/proto"
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"
)

func (s *AdminService) UpdateUserRole(ctx context.Context, id int, isAdmin bool) (domain.User, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		auth := md.Get("authorization")
		if len(auth) > 0 {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", auth[0])
		}
	}
	if id <= 0 {
		return domain.User{}, fmt.Errorf("Id must be greater than 0")
	}
	grpc := &userpb.UpdateRoleRequest{
		Id:      int32(id),
		IsAdmin: isAdmin,
	}
	update, err := s.userClient.UpdateRole(ctx, grpc)
	if err != nil {
		return domain.User{}, fmt.Errorf("Update User Role: %w", err)
	}

	return domain.User{
		ID:          update.ID,
		FullName:    update.FullName,
		Email:       update.Email,
		PhoneNumber: update.PhoneNumber,
		IsAdmin:     update.IsAdmin,
	}, nil
}
