package service_admin

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"

	"github.com/daniiiiiiiiiiil/finance-microservices/admin-service/internal/core/domain"
	userpb "github.com/daniiiiiiiiiiil/finance-microservices/admin-service/proto/users/gen"
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
