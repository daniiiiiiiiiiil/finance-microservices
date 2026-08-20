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
		Version:     int32(user.Version),
		FullName:    user.FullName,
		Email:       user.Email,
		PhoneNumber: phoneNumber,
		IsAdmin:     user.IsAdmin,
		IsActive:    user.Status == "active",
		Status:      user.Status,
	}
}

func convertRegisterRequestToDomain(req *proto.RegisterRequest) domain.User {
	var phoneNumber *string
	if req.PhoneNumber != "" {
		phoneNumber = &req.PhoneNumber
	}
	return domain.NewUserUninitialized(
		req.FullName,
		req.Email,
		"",
		phoneNumber,
		req.IsAdmin,
		"pending",
	)
}

func convertLoginRequestToDomain(req *proto.LoginRequest) (string, string) {
	return req.Email, req.Password
}
