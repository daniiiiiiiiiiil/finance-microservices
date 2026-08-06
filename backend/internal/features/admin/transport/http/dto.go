package http

import "backend/internal/core/domain"

type AdminUserResponse struct {
	ID          int     `json:"id"`
	Version     int     `json:"version"`
	FullName    string  `json:"full_name"`
	Email       string  `json:"email"`
	PhoneNumber *string `json:"phone_number"`
	IsAdmin     bool    `json:"is_admin"`
}

type GetUsersResponse struct {
	Data  []AdminUserResponse `json:"data"`
	Total int                 `json:"total"`
	Limit int                 `json:"limit"`
	Page  int                 `json:"page"`
}

type UpdateRoleRequest struct {
	IsAdmin bool `json:"is_admin"`
}

type MetricsResponse struct {
	TotalUsers        int     `json:"total_users"`
	TotalTransactions int     `json:"total_transactions"`
	TotalBalance      float64 `json:"total_balance"`
}

func adminUserResponseFromDomain(user domain.User) AdminUserResponse {
	return AdminUserResponse{
		ID:          user.ID,
		Version:     user.Version,
		FullName:    user.FullName,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		IsAdmin:     user.IsAdmin,
	}
}

func adminUserResponsesFromDomains(users []domain.User) []AdminUserResponse {
	responses := make([]AdminUserResponse, len(users))
	for i, user := range users {
		responses[i] = adminUserResponseFromDomain(user)
	}
	return responses
}
