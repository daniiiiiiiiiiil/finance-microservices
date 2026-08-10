package grpc

import (
	"backend/internal/core/domain"
)

func convertUserToProto(user domain.User) *UserResponse {
	var phoneNumber *string
	if user.PhoneNumber != nil && *user.PhoneNumber != "" {
		phoneNumber = &(*user.PhoneNumber)
	}
	return &UserResponse{
		Id:          int32(user.ID),
		Version:     int32(user.Version),
		FullName:    user.FullName,
		Email:       user.Email,
		PhoneNumber: phoneNumber,
		IsAdmin:     user.IsAdmin,
	}
}

func convertPatchProtoToDomain(data *PatchUserData) domain.UserPatch {
	return domain.NewUserPatch(
		convertOptionalStringToNullable(data.FullName),
		convertOptionalStringToNullable(data.PhoneNumber),
	)
}

func convertOptionalStringToNullable(value *string) domain.Nullable[string] {
	if value == nil {
		return domain.Nullable[string]{Value: nil, Set: false}
	}
	return domain.Nullable[string]{Value: value, Set: true}
}

func ConvertUsersToProto(users []domain.User) []*UserResponse {
	result := make([]*UserResponse, len(users))
	for i, user := range users {
		result[i] = convertUserToProto(user)
	}
	return result
}
