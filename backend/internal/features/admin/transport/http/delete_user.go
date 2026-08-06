package http

import (
	"backend/internal/core/logger"
	core_middleware "backend/internal/core/transport/http/middleware"
	"backend/internal/core/transport/http/request"
	response_core "backend/internal/core/transport/http/response"
	"errors"
	"net/http"
)

// DeleteUser удаляет пользователя (только для админов)
// @Summary Удалить пользователя
// @Description Удаляет пользователя по ID. Доступно только администраторам. Транзакции пользователя удаляются автоматически (CASCADE).
// @Tags admin
// @Produce json
// @Param id path int true "ID пользователя"
// @Success 204 "Пользователь успешно удалён"
// @Failure 400 {object} response_core.ErrorResponse "Неверный ID пользователя"
// @Failure 401 {object} response_core.ErrorResponse "Не авторизован"
// @Failure 403 {object} response_core.ErrorResponse "Доступ запрещён (не админ или удаление себя)"
// @Failure 404 {object} response_core.ErrorResponse "Пользователь не найден"
// @Failure 500 {object} response_core.ErrorResponse "Внутренняя ошибка сервера"
// @Router /admin/users/{id} [delete]
func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
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
		rh.ErrorResponse(errors.New("forbidden"), "cannot delete yourself")
		return
	}

	if err := h.service.DeleteUser(ctx, id); err != nil {
		rh.ErrorResponse(err, "failed to delete user")
		return
	}

	rh.NoContentResponse()
}
