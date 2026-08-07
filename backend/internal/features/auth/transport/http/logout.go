package http_auth

import (
	"fmt"
	"net/http"
	"time"
)

// Logout godoc
// @Summary Выход из системы
// @Description Удаляет JWT токен из cookie
// @Tags auth
// @Success 204 "Успешный выход"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cookie, err := r.Cookie("token")
	if err == nil && cookie != nil {
		if err := h.service.AddToBlacklist(ctx, cookie.Value, 24*time.Hour); err != nil {
			fmt.Printf("failed to add token to blacklist: %v\n", err)
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusNoContent)
}
