package service_admin

import (
	"backend/internal/core/domain"
	userpb "backend/internal/features/users/transport/grpc/proto"
	"context"
	"fmt"
)

func (s *AdminService) UpdateUserRole(ctx context.Context, id int, isAdmin bool) (domain.User, error) {
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
