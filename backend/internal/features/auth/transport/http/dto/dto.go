package http_auth

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	FullName    string `json:"full_name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	PhoneNumber string `json:"phone_number"`
	IsAdmin     bool   `json:"is_admin"`
}

type UserResponse struct {
	ID          int     `json:"id"`
	FullName    string  `json:"full_name"`
	Email       string  `json:"email"`
	PhoneNumber *string `json:"phone_number"`
	IsAdmin     bool    `json:"is_admin"`
}
