package http_auth

import (
	"backend/internal/core/logger"
	"backend/internal/core/transport/http/request"
	response_core "backend/internal/core/transport/http/response"
	service_auth "backend/internal/features/auth/service"
	http_auth "backend/internal/features/auth/transport/http/dto"
	"errors"
	"net/http"
)

// Register godoc
// @Summary Регистрация пользователя
// @Description Создаёт нового пользователя и сохраняет JWT токен в HttpOnly cookie
// @Tags auth
// @Accept json
// @Produce json
// @Param request body http_auth.RegisterRequest true "Данные для регистрации"
// @Success 201 {object} http_auth.UserResponse
// @Failure 400 {object} response_core.ErrorResponse
// @Failure 409 {object} response_core.ErrorResponse
// @Failure 500 {object} response_core.ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	rh := response_core.NewHTTPResponseHandler(log, w)

	var req http_auth.RegisterRequest
	if err := request.DecodeAndValidateRequest(r, &req); err != nil {
		rh.ErrorResponse(err, "invalid request")
		return
	}

	token, user, err := h.service.Register(ctx, req)
	if err != nil {
		switch {
		case errors.Is(err, service_auth.ErrUserAlreadyExists):
			rh.ErrorResponse(err, "user already exists")
		default:
			rh.ErrorResponse(err, "registration failed")
		}
		return
	}

	setAuthCookie(w, token)

	rh.JSONResponse(user, http.StatusCreated)
}
