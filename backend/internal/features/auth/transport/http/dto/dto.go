package dto

type RegisterRequest struct {
	FullName    string `json:"full_name" validate:"required,min=3,max=100"`
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
	PhoneNumber string `json:"phone_number" validate:"omitempty,min=10,max=15"`
	IsAdmin     bool   `json:"is_admin"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type UserResponse struct {
	ID          int     `json:"id"`
	FullName    string  `json:"full_name"`
	Email       string  `json:"email"`
	PhoneNumber *string `json:"phone_number"`
	IsAdmin     bool    `json:"is_admin"`
}
