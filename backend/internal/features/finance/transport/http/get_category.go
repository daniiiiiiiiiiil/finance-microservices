package http

import (
	"backend/internal/core/logger"
	core_middleware "backend/internal/core/transport/http/middleware"
	response_core "backend/internal/core/transport/http/response"
	"errors"
	"net/http"
)

// GetCategories возвращает все категории пользователя
// @Summary Получить категории
// @Description Возвращает все уникальные категории транзакций пользователя
// @Tags finance
// @Produce json
// @Success 200 {array} string "Список категорий"
// @Failure 401 {object} response_core.ErrorResponse "Не авторизован"
// @Failure 500 {object} response_core.ErrorResponse "Внутренняя ошибка сервера"
// @Router /finance/categories [get]
func (h *TransactionHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := response_core.NewHTTPResponseHandler(log, w)
	userID, ok := core_middleware.GetUserID(r)
	if !ok {
		responseHandler.ErrorResponse(errors.New("unauthorized"), "user not authenticated")
		return
	}
	categories, err := h.service.GetCategories(ctx, userID)
	if err != nil {
		responseHandler.ErrorResponse(errors.New("unauthorized"), "user not authenticated")
	}
	responseHandler.JSONResponse(categories, http.StatusOK)
}
