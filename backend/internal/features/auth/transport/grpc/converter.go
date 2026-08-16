package grpc

import (
	"backend/internal/core/domain"
	"backend/internal/features/auth/transport/grpc/proto"
)

func convertUserToProto(user domain.User) *proto.UserResponse {
	var phoneNumber *string
	if user.PhoneNumber != nil && *user.PhoneNumber != "" {
		phoneNumber = user.PhoneNumber
	}
	return &proto.UserResponse{
		Id:          int32(user.ID),
		FullName:    user.FullName,
		Email:       user.Email,
		PhoneNumber: phoneNumber,
		IsAdmin:     user.IsAdmin,
	}
}

func convertRegisterRequestToDomain(req *proto.RegisterRequest) domain.User {
	return domain.NewUserUninitialized(
		req.FullName,
		req.Email,
		"",
		&req.PhoneNumber,
		req.IsAdmin,
	)
}

func convertLoginRequestToDomain(req *proto.LoginRequest) (string, string) {
	return req.Email, req.Password
}
