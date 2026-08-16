package http_auth

import (
	"backend/internal/core/logger"
	"backend/internal/core/transport/http/request"
	response_core "backend/internal/core/transport/http/response"
	service_auth "backend/internal/features/auth/service"
	"backend/internal/features/auth/transport/http/dto"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Login godoc
// @Summary Авторизация пользователя
// @Description Авторизует пользователя и сохраняет JWT токен в HttpOnly cookie
// @Tags auth
// @Accept json
// @Produce json
// @Param request body http_auth.LoginRequest true "Данные для входа"
// @Success 200 {object} http_auth.UserResponse
// @Failure 400 {object} response_core.ErrorResponse
// @Failure 401 {object} response_core.ErrorResponse
// @Failure 500 {object} response_core.ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	rh := response_core.NewHTTPResponseHandler(log, w)

	ip := r.RemoteAddr
	key := fmt.Sprintf("rate:login:%s", ip)

	allowed, err := h.service.RateLimitCheck(ctx, key, 5, 1*time.Minute)
	if err != nil {
		rh.ErrorResponse(err, "rate limit check failed")
		return
	}
	if !allowed {
		rh.ErrorResponse(errors.New("too many attempts"), "try again later")
		return
	}

	var req http_auth.LoginRequest
	if err := request.DecodeAndValidateRequest(r, &req); err != nil {
		rh.ErrorResponse(err, "invalid request")
		return
	}

	token, user, err := h.service.Login(ctx, req)
	if err != nil {
		if errors.Is(err, service_auth.ErrInvalidCredentials) {
			rh.ErrorResponse(err, "invalid credentials")
			return
		}
		rh.ErrorResponse(err, "login failed")
		return
	}

	setAuthCookie(w, token)

	rh.JSONResponse(user, http.StatusOK)
}
