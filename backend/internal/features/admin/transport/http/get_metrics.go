package http

import (
	"backend/internal/core/logger"
	response_core "backend/internal/core/transport/http/response"
	"errors"
	"net/http"
)

// GetMetrics возвращает системные метрики
// @Summary Получить системные метрики
// @Description Возвращает общую статистику системы: количество пользователей, количество транзакций и общий баланс. Доступно только администраторам.
// @Tags admin
// @Produce json
// @Success 200 {object} MetricsResponse "Системные метрики"
// @Failure 401 {object} response_core.ErrorResponse "Не авторизован"
// @Failure 403 {object} response_core.ErrorResponse "Доступ запрещён (требуются права администратора)"
// @Failure 500 {object} response_core.ErrorResponse "Внутренняя ошибка сервера"
// @Router /admin/metrics [get]
func (h *AdminHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	rh := response_core.NewHTTPResponseHandler(log, w)

	if !h.isAdmin(r) {
		rh.ErrorResponse(errors.New("forbidden"), "admin access required")
		return
	}

	metrics, err := h.service.GetMetrics(ctx)
	if err != nil {
		rh.ErrorResponse(err, "failed to get metrics")
		return
	}

	rh.JSONResponse(metrics, http.StatusOK)
}
