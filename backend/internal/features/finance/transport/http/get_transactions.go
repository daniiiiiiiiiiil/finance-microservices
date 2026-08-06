package http

import (
	"backend/internal/core/logger"
	core_middleware "backend/internal/core/transport/http/middleware"
	response_core "backend/internal/core/transport/http/response"
	"errors"
	"net/http"
	"strconv"
	"time"
)

// GetTransactions возвращает список транзакций с фильтрацией
// @Summary Получить список транзакций
// @Description Возвращает транзакции пользователя с фильтрацией и пагинацией
// @Tags finance
// @Produce json
// @Param type query string false "Тип транзакции (income/expense)"
// @Param category query string false "Категория"
// @Param from query string false "Дата начала (YYYY-MM-DD)"
// @Param to query string false "Дата конца (YYYY-MM-DD)"
// @Param limit query int false "Лимит (default 20)"
// @Param offset query int false "Смещение (default 0)"
// @Success 200 {array} TransactionResponse
// @Failure 400 {object} response_core.ErrorResponse
// @Failure 401 {object} response_core.ErrorResponse
// @Failure 500 {object} response_core.ErrorResponse
// @Router /finance/transactions [get]
func (h *TransactionHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	rh := response_core.NewHTTPResponseHandler(log, w)

	userID, ok := core_middleware.GetUserID(r)
	if !ok {
		rh.ErrorResponse(errors.New("unauthorized"), "user not authenticated")
		return
	}

	transactionType := r.URL.Query().Get("type")
	if transactionType != "" && transactionType != "income" && transactionType != "expense" {
		rh.ErrorResponse(errors.New("invalid type"), "type must be 'income' or 'expense'")
		return
	}

	category := r.URL.Query().Get("category")

	fromStr := r.URL.Query().Get("from")
	var from *time.Time
	if fromStr != "" {
		t, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			rh.ErrorResponse(err, "invalid 'from' date format (use YYYY-MM-DD)")
			return
		}
		from = &t
	}

	toStr := r.URL.Query().Get("to")
	var to *time.Time
	if toStr != "" {
		t, err := time.Parse("2006-01-02", toStr)
		if err != nil {
			rh.ErrorResponse(err, "invalid 'to' date format (use YYYY-MM-DD)")
			return
		}
		t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		to = &t
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	var typePtr *string
	if transactionType != "" {
		typePtr = &transactionType
	}

	var categoryPtr *string
	if category != "" {
		categoryPtr = &category
	}

	transactions, err := h.service.GetTransactions(ctx, userID, typePtr, categoryPtr, from, to, limit, offset)
	if err != nil {
		rh.ErrorResponse(err, "failed to get transactions")
		return
	}

	response := transactionResponsesFromDomains(transactions)
	rh.JSONResponse(response, http.StatusOK)
}
