package gRPC

import (
	"backend/internal/core/domain"
	"backend/internal/features/admin/transport/gRPC/proto"
)

func convertUserToProto(user domain.User) *proto.AdminUserResponse {
	var phoneNumber *string
	if user.PhoneNumber != nil && *user.PhoneNumber != "" {
		phoneNumber = user.PhoneNumber
	}
	return &proto.AdminUserResponse{
		Id:          int32(user.ID),
		Version:     int32(user.Version),
		FullName:    user.FullName,
		Email:       user.Email,
		PhoneNumber: phoneNumber,
		IsAdmin:     user.IsAdmin,
	}
}

func convertUsersToProto(users []domain.User, limit, offset int) *proto.GetUsersResponse {
	responses := make([]*proto.AdminUserResponse, len(users))
	for i, user := range users {
		responses[i] = convertUserToProto(user)
	}
	return &proto.GetUsersResponse{
		Data:  responses,
		Limit: int32(limit),
		Page:  int32(offset/limit + 1),
	}
}
