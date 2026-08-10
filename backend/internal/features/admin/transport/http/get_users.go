package http

import (
	"backend/internal/core/logger"
	"backend/internal/core/pagination"
	response_core "backend/internal/core/transport/http/response"
	"errors"
	"net/http"
	"strconv"
)

// GetUsers возвращает список всех пользователей с пагинацией
// @Summary Получить список пользователей
// @Description Возвращает список всех пользователей с пагинацией. Доступно только администраторам.
// @Tags admin
// @Produce json
// @Param limit query int false "Лимит (default 20)"
// @Param offset query int false "Смещение (default 0)"
// @Success 200 {object} GetUsersResponse "Список пользователей с пагинацией"
// @Failure 401 {object} response_core.ErrorResponse "Не авторизован"
// @Failure 403 {object} response_core.ErrorResponse "Доступ запрещён (требуются права администратора)"
// @Failure 500 {object} response_core.ErrorResponse "Внутренняя ошибка сервера"
// @Router /admin/users [get]
func (h *AdminHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	rh := response_core.NewHTTPResponseHandler(log, w)

	if !h.isAdmin(r) {
		rh.ErrorResponse(errors.New("forbidden"), "admin access required")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	limit, offset = pagination.LimitOffset(limit, offset)

	users, total, err := h.service.GetUsers(ctx, limit, offset)
	if err != nil {
		rh.ErrorResponse(err, "failed to get users")
		return
	}

	pag := pagination.NewPagination(total, limit, offset)

	response := struct {
		Data       []AdminUserResponse   `json:"data"`
		Pagination pagination.Pagination `json:"pagination"`
	}{
		Data:       adminUserResponsesFromDomains(users),
		Pagination: pag,
	}

	rh.JSONResponse(response, http.StatusOK)
}
