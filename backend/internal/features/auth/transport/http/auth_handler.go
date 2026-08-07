package http_auth

import (
	"backend/internal/features/auth/service"
	"net/http"
)

type AuthHandler struct {
	service *service_auth.AuthService
}

func NewAuthHandler(service *service_auth.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func setAuthCookie(w http.ResponseWriter, token string) {
	maxAge := 86400

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}
