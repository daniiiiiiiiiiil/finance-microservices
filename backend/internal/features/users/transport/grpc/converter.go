package grpc

import (
	"backend/internal/core/domain"
	"backend/internal/features/users/transport/grpc/proto"
)

func convertUserToProto(user domain.User) *proto.UserResponse {
	phoneNumber := ""
	if user.PhoneNumber != nil && *user.PhoneNumber != "" {
		phoneNumber = *user.PhoneNumber
	}
	return &proto.UserResponse{
		Id:          int32(user.ID),
		Version:     int32(user.Version),
		FullName:    user.FullName,
		Email:       user.Email,
		PhoneNumber: phoneNumber,
		IsAdmin:     user.IsAdmin,
	}
}

func convertPatchProtoToDomain(data *proto.PatchUserData) domain.UserPatch {
	return domain.NewUserPatch(
		convertStringToNullable(data.FullName),
		convertStringToNullable(data.PhoneNumber),
	)
}

func convertStringToNullable(value string) domain.Nullable[string] {
	if value == "" {
		return domain.Nullable[string]{Value: nil, Set: false}
	}
	return domain.Nullable[string]{Value: &value, Set: true}
}

func ConvertUsersToProto(users []domain.User) []*proto.UserResponse {
	result := make([]*proto.UserResponse, len(users))
	for i, user := range users {
		result[i] = convertUserToProto(user)
	}
	return result
}
