package grpc

import (
	"backend/internal/core/domain"
	"backend/proto/users/gen"
)

func convertUserToProto(user domain.User) *gen.UserResponse {
	phoneNumber := ""
	if user.PhoneNumber != nil && *user.PhoneNumber != "" {
		phoneNumber = *user.PhoneNumber
	}
	return &gen.UserResponse{
		Id:          int32(user.ID),
		Version:     int32(user.Version),
		FullName:    user.FullName,
		Email:       user.Email,
		PhoneNumber: phoneNumber,
		IsAdmin:     user.IsAdmin,
	}
}

func convertPatchProtoToDomain(data *gen.PatchUserData) domain.UserPatch {
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

func ConvertUsersToProto(users []domain.User) []*gen.UserResponse {
	result := make([]*gen.UserResponse, len(users))
	for i, user := range users {
		result[i] = convertUserToProto(user)
	}
	return result
}
