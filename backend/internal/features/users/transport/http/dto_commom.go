package user_transport_http

import "backend/internal/core/domain"

type UserDTOResponse struct {
	ID          int     `json:"id" example:"10"`
	Version     int     `json:"version" example:"1"`
	FullName    string  `json:"full_name" example:"FirstName LastName"`
	Email       string  `json:"email" example:"email@gmain.com"`
	PhoneNumber *string `json:"phone_number" example:"+799999999"`
	IsAdmin     bool    `json:"is_admin" example:"false"`
}

func UserDTOFromDomain(user domain.User) UserDTOResponse {
	return UserDTOResponse{
		ID:          user.ID,
		Version:     user.Version,
		FullName:    user.FullName,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		IsAdmin:     user.IsAdmin,
	}
}

func usersDTOFromDomains(users []domain.User) []UserDTOResponse {
	userDTOs := make([]UserDTOResponse, len(users))
	for i, user := range users {
		userDTOs[i] = UserDTOFromDomain(user)
	}
	return userDTOs
}
