package http

import (
	"backend/internal/core/domain"
	"backend/internal/core/logger"
	core_middleware "backend/internal/core/transport/http/middleware"
	"backend/internal/core/transport/http/request"
	response_core "backend/internal/core/transport/http/response"
	"errors"
	"net/http"
	"time"
)

// CreateTransaction создаёт новую транзакцию
// @Summary Создать транзакцию
// @Description Создаёт новую транзакцию (доход или расход)
// @Tags finance
// @Accept json
// @Produce json
// @Param request body CreateTransactionRequest true "Данные транзакции"
// @Success 201 {object} TransactionResponse
// @Failure 400 {object} response_core.ErrorResponse
// @Failure 401 {object} response_core.ErrorResponse
// @Failure 500 {object} response_core.ErrorResponse
// @Router /finance/transactions [post]
func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	rh := response_core.NewHTTPResponseHandler(log, w)

	// Получаем user_id из контекста
	userID, ok := core_middleware.GetUserID(r)
	if !ok {
		rh.ErrorResponse(errors.New("unauthorized"), "user not authenticated")
		return
	}

	var req CreateTransactionRequest
	if err := request.DecodeAndValidateRequest(r, &req); err != nil {
		rh.ErrorResponse(err, "invalid request")
		return
	}

	transaction := domain.Finance{
		ID:              -1,
		Version:         -1,
		TypeTransaction: req.TypeTransaction,
		Amount:          req.Amount,
		Category:        req.Category,
		CreatedAt:       time.Now(),
		UserID:          userID,
	}

	created, err := h.service.CreateTransaction(ctx, transaction)
	if err != nil {
		rh.ErrorResponse(err, "failed to create transaction")
		return
	}

	response := transactionResponseFromDomain(created)
	rh.JSONResponse(response, http.StatusCreated)
}
