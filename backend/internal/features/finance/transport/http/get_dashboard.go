// internal/features/finance/transport/http/dashboard_handler.go
package http

import (
	"backend/internal/core/domain"
	"backend/internal/core/logger"
	core_middleware "backend/internal/core/transport/http/middleware"
	"backend/internal/core/transport/http/response"
	"errors"
	"net/http"
)

type DashboardResponse = domain.Dashboard

// GetDashboard возвращает дашборд пользователя
// @Summary Получить дашборд
// @Description Возвращает финансовый дашборд с метриками, графиком и последними операциями
// @Tags finance
// @Produce json
// @Success 200 {object} DashboardResponse
// @Failure 401 {object} response_core.ErrorResponse
// @Failure 500 {object} response_core.ErrorResponse
// @Router /finance/dashboard [get]
func (h *TransactionHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := response_core.NewHTTPResponseHandler(log, w)

	userID, ok := core_middleware.GetUserID(r)
	if !ok {
		responseHandler.ErrorResponse(errors.New("unauthorized"), "user not authenticated")
		return
	}

	dashboard, err := h.service.GetDashboard(ctx, userID)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get dashboard")
		return
	}

	responseHandler.JSONResponse(dashboard, http.StatusOK)
}
