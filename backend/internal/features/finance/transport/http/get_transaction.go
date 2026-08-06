package http

import (
	"backend/internal/core/logger"
	core_middleware "backend/internal/core/transport/http/middleware"
	"backend/internal/core/transport/http/request"
	response_core "backend/internal/core/transport/http/response"
	"errors"
	"net/http"
)

// GetTransaction возвращает транзакцию по ID
// @Summary Получить транзакцию
// @Description Возвращает детали транзакции по ID
// @Tags finance
// @Produce json
// @Param id path int true "ID транзакции"
// @Success 200 {object} TransactionResponse
// @Failure 400 {object} response_core.ErrorResponse
// @Failure 401 {object} response_core.ErrorResponse
// @Failure 404 {object} response_core.ErrorResponse
// @Failure 500 {object} response_core.ErrorResponse
// @Router /finance/transactions/{id} [get]
func (h *TransactionHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
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
		rh.ErrorResponse(errors.New("forbidden"), "you can only view your own transactions")
		return
	}

	response := transactionResponseFromDomain(transaction)
	rh.JSONResponse(response, http.StatusOK)
}
