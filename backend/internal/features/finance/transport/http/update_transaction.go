package http

import (
	"backend/internal/core/domain"
	"backend/internal/core/logger"
	core_middleware "backend/internal/core/transport/http/middleware"
	"backend/internal/core/transport/http/request"
	response_core "backend/internal/core/transport/http/response"
	"errors"
	"net/http"
)

// UpdateTransaction обновляет транзакцию
// @Summary Обновить транзакцию
// @Description Обновляет существующую транзакцию
// @Tags finance
// @Accept json
// @Produce json
// @Param id path int true "ID транзакции"
// @Param request body UpdateTransactionRequest true "Данные для обновления"
// @Success 200 {object} TransactionResponse
// @Failure 400 {object} response_core.ErrorResponse
// @Failure 401 {object} response_core.ErrorResponse
// @Failure 404 {object} response_core.ErrorResponse
// @Failure 409 {object} response_core.ErrorResponse
// @Failure 500 {object} response_core.ErrorResponse
// @Router /finance/transactions/{id} [put]
func (h *TransactionHandler) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
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

	var req UpdateTransactionRequest
	if err := request.DecodeAndValidateRequest(r, &req); err != nil {
		rh.ErrorResponse(err, "invalid request")
		return
	}

	existing, err := h.service.GetTransaction(ctx, id)
	if err != nil {
		rh.ErrorResponse(err, "failed to get transaction")
		return
	}

	if existing.UserID != userID {
		rh.ErrorResponse(errors.New("forbidden"), "you can only update your own transactions")
		return
	}

	updated := domain.Finance{
		ID:              id,
		Version:         existing.Version,
		TypeTransaction: req.TypeTransaction,
		Amount:          req.Amount,
		Category:        req.Category,
		CreatedAt:       existing.CreatedAt,
		UserID:          userID,
	}

	result, err := h.service.UpdateTransaction(ctx, updated)
	if err != nil {
		rh.ErrorResponse(err, "failed to update transaction")
		return
	}

	response := transactionResponseFromDomain(result)
	rh.JSONResponse(response, http.StatusOK)
}
