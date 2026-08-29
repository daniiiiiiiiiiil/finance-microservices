package grpc

import (
	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/internal/core/domain"
	"github.com/daniiiiiiiiiiil/finance-microservices/auth-service/proto/auth/gen"
)

func convertUserToProto(user domain.User) *gen.UserResponse {
	var phoneNumber *string
	if user.PhoneNumber != nil && *user.PhoneNumber != "" {
		phoneNumber = user.PhoneNumber
	}
	return &gen.UserResponse{
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

//func convertRegisterRequestToDomain(req *gen.RegisterRequest) domain.User {
//	var phoneNumber *string
//	if req.PhoneNumber != "" {
//		phoneNumber = &req.PhoneNumber
//	}
//	return domain.NewUserUninitialized(
//		req.FullName,
//		req.Email,
//		"",
//		phoneNumber,
//		req.IsAdmin,
//		"pending",
//	)
//}
//
//func convertLoginRequestToDomain(req *gen.LoginRequest) (string, string) {
//	return req.Email, req.Password
//}
