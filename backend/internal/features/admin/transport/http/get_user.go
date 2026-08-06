package http

import (
	"backend/internal/core/logger"
	"backend/internal/core/transport/http/request"
	response_core "backend/internal/core/transport/http/response"
	"errors"
	"net/http"
)

// GetUser возвращает данные пользователя по ID
// @Summary Получить пользователя
// @Description Возвращает данные пользователя по ID. Доступно только администраторам.
// @Tags admin
// @Produce json
// @Param id path int true "ID пользователя"
// @Success 200 {object} AdminUserResponse "Данные пользователя"
// @Failure 400 {object} response_core.ErrorResponse "Неверный ID пользователя"
// @Failure 401 {object} response_core.ErrorResponse "Не авторизован"
// @Failure 403 {object} response_core.ErrorResponse "Доступ запрещён (требуются права администратора)"
// @Failure 404 {object} response_core.ErrorResponse "Пользователь не найден"
// @Failure 500 {object} response_core.ErrorResponse "Внутренняя ошибка сервера"
// @Router /admin/users/{id} [get]
func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	rh := response_core.NewHTTPResponseHandler(log, w)

	if !h.isAdmin(r) {
		rh.ErrorResponse(errors.New("forbidden"), "admin access required")
		return
	}

	id, err := request.GetIntPathValue(r, "id")
	if err != nil {
		rh.ErrorResponse(err, "invalid user id")
		return
	}

	user, err := h.service.GetUser(ctx, id)
	if err != nil {
		rh.ErrorResponse(err, "failed to get user")
		return
	}

	rh.JSONResponse(adminUserResponseFromDomain(user), http.StatusOK)
}
