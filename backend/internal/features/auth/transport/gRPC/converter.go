package gRPC

import (
	"backend/internal/core/domain"
)

func convertUserToProto(user domain.User) *UserResponse {
	var phoneNumber *string
	if user.PhoneNumber != nil && *user.PhoneNumber != "" {
		phoneNumber = user.PhoneNumber
	}
	return &UserResponse{
		Id:          int32(user.ID),
		FullName:    user.FullName,
		Email:       user.Email,
		PhoneNumber: phoneNumber,
		IsAdmin:     user.IsAdmin,
	}
}

func convertRegisterRequestToDomain(req *RegisterRequest) domain.User {
	return domain.NewUserUninitialized(
		req.FullName,
		req.Email,
		"",
		&req.PhoneNumber,
		req.IsAdmin,
	)
}

func convertLoginRequestToDomain(req *LoginRequest) (string, string) {
	return req.Email, req.Password
}
