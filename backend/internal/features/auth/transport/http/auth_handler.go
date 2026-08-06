package http_auth

import (
	"backend/internal/core/logger"
	"backend/internal/core/transport/http/request"
	"backend/internal/core/transport/http/response"
	"backend/internal/features/auth/service"
	"backend/internal/features/auth/transport/http/dto"
	"errors"
	"net/http"
)

type AuthHandler struct {
	service *service_auth.AuthService
}

func NewAuthHandler(service *service_auth.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// Register godoc
// @Summary Регистрация пользователя
// @Description Создаёт нового пользователя и сохраняет JWT токен в HttpOnly cookie
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Данные для регистрации"
// @Success 201 {object} dto.UserResponse
// @Failure 400 {object} response_core.ErrorResponse
// @Failure 409 {object} response_core.ErrorResponse
// @Failure 500 {object} response_core.ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	rh := response_core.NewHTTPResponseHandler(log, w)

	var req dto.RegisterRequest
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

// Login godoc
// @Summary Авторизация пользователя
// @Description Авторизует пользователя и сохраняет JWT токен в HttpOnly cookie
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Данные для входа"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} response_core.ErrorResponse
// @Failure 401 {object} response_core.ErrorResponse
// @Failure 500 {object} response_core.ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	rh := response_core.NewHTTPResponseHandler(log, w)

	var req dto.LoginRequest
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

// Logout godoc
// @Summary Выход из системы
// @Description Удаляет JWT токен из cookie
// @Tags auth
// @Success 204 "Успешный выход"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
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
