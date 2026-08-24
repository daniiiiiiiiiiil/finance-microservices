package grpc

import (
	"backend/internal/core/domain"
	"backend/proto/admin/gen"
)

func convertUserToProto(user domain.User) *gen.AdminUserResponse {
	var phoneNumber *string
	if user.PhoneNumber != nil && *user.PhoneNumber != "" {
		phoneNumber = user.PhoneNumber
	}
	return &gen.AdminUserResponse{
		Id:          int32(user.ID),
		Version:     int32(user.Version),
		FullName:    user.FullName,
		Email:       user.Email,
		PhoneNumber: phoneNumber,
		IsAdmin:     user.IsAdmin,
	}
}

func convertUsersToProto(users []domain.User, limit, offset int) *gen.GetUsersResponse {
	responses := make([]*gen.AdminUserResponse, len(users))
	for i, user := range users {
		responses[i] = convertUserToProto(user)
	}
	return &gen.GetUsersResponse{
		Data:  responses,
		Limit: int32(limit),
		Page:  int32(offset/limit + 1),
	}
}
