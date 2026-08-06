package http

import (
	"backend/internal/core/logger"
	core_middleware "backend/internal/core/transport/http/middleware"
	"backend/internal/core/transport/http/request"
	response_core "backend/internal/core/transport/http/response"
	"errors"
	"net/http"
)

// UpdateUserRole изменяет роль пользователя
// @Summary Изменить роль пользователя
// @Description Назначает или снимает права администратора. Доступно только администраторам. Нельзя изменить свою собственную роль.
// @Tags admin
// @Accept json
// @Produce json
// @Param id path int true "ID пользователя"
// @Param request body UpdateRoleRequest true "Новая роль"
// @Success 200 {object} AdminUserResponse "Обновлённые данные пользователя"
// @Failure 400 {object} response_core.ErrorResponse "Неверный ID пользователя или тело запроса"
// @Failure 401 {object} response_core.ErrorResponse "Не авторизован"
// @Failure 403 {object} response_core.ErrorResponse "Доступ запрещён (не админ или изменение своей роли)"
// @Failure 404 {object} response_core.ErrorResponse "Пользователь не найден"
// @Failure 500 {object} response_core.ErrorResponse "Внутренняя ошибка сервера"
// @Router /admin/users/{id}/role [patch]
func (h *AdminHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
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

	currentUserID, _ := core_middleware.GetUserID(r)
	if id == currentUserID {
		rh.ErrorResponse(errors.New("forbidden"), "cannot change your own role")
		return
	}

	var req UpdateRoleRequest
	if err := request.DecodeAndValidateRequest(r, &req); err != nil {
		rh.ErrorResponse(err, "invalid request")
		return
	}

	user, err := h.service.UpdateUserRole(ctx, id, req.IsAdmin)
	if err != nil {
		rh.ErrorResponse(err, "failed to update user role")
		return
	}

	rh.JSONResponse(adminUserResponseFromDomain(user), http.StatusOK)
}
