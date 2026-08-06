package http

import (
	"backend/internal/core/logger"
	core_middleware "backend/internal/core/transport/http/middleware"
	"backend/internal/core/transport/http/request"
	response_core "backend/internal/core/transport/http/response"
	"errors"
	"net/http"
)

// DeleteTransaction удаляет транзакцию
// @Summary Удалить транзакцию
// @Description Удаляет транзакцию по ID
// @Tags finance
// @Produce json
// @Param id path int true "ID транзакции"
// @Success 204 "Успешно удалено"
// @Failure 400 {object} response_core.ErrorResponse
// @Failure 401 {object} response_core.ErrorResponse
// @Failure 404 {object} response_core.ErrorResponse
// @Failure 500 {object} response_core.ErrorResponse
// @Router /finance/transactions/{id} [delete]
func (h *TransactionHandler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	rh := response_core.NewHTTPResponseHandler(log, w)

	userID, ok := core_middleware.GetUserID(r)
	if !ok {
		rh.ErrorResponse(errors.New("unauthorized"), "user not authenticated")
		return
	}

	id, err := request.GetIntPathValue(r, "id")
	if err != nil {
		rh.ErrorResponse(err, "invalid transaction id")
		return
	}

	transaction, err := h.service.GetTransaction(ctx, id)
	if err != nil {
		rh.ErrorResponse(err, "failed to get transaction")
		return
	}

	if transaction.UserID != userID {
		rh.ErrorResponse(errors.New("forbidden"), "you can only delete your own transactions")
		return
	}

	if err := h.service.DeleteTransaction(ctx, id); err != nil {
		rh.ErrorResponse(err, "failed to delete transaction")
		return
	}

	rh.NoContentResponse()
}
