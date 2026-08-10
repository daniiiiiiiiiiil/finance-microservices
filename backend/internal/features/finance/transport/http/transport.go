package http

import (
	"backend/internal/core/auth/jwt"
	"backend/internal/core/domain"
	core_middleware "backend/internal/core/transport/http/middleware"
	"backend/internal/core/transport/http/server"
	"backend/internal/features/finance/service"
	"context"
	"net/http"
	"time"
)

type TransactionHandler struct {
	service *service.FinanceService
}

func NewTransactionHandler(service *service.FinanceService) *TransactionHandler {
	return &TransactionHandler{service: service}
}

type FinanceService interface {
	GetDashboard(ctx context.Context, userID int) (domain.Dashboard, error)
	GetCategories(ctx context.Context, userID int) ([]string, error)
	CreateTransaction(ctx context.Context, transaction domain.Finance) (domain.Finance, error)
	GetTransaction(ctx context.Context, id int) (domain.Finance, error)
	GetTransactions(ctx context.Context, userID int, transactionType, category *string, from, to *time.Time, limit, offset int) ([]domain.Finance, error)
	UpdateTransaction(ctx context.Context, transaction domain.Finance) (domain.Finance, error)
	DeleteTransaction(ctx context.Context, id int) error
}

type FinanceHandler struct {
	transactionHandler *TransactionHandler
	jwtManager         *jwt.JWTManager
}

func NewFinanceHandler(dashboardService *service.FinanceService, jwtManager *jwt.JWTManager) *FinanceHandler {
	return &FinanceHandler{
		transactionHandler: NewTransactionHandler(dashboardService),
		jwtManager:         jwtManager,
	}
}

func (h *FinanceHandler) Routes() []server.Route {
	authMiddleware := []core_middleware.Middleware{
		core_middleware.Auth(h.jwtManager),
	}

	return []server.Route{
		{
			Method:     http.MethodGet,
			Path:       "/finance/dashboard",
			Handler:    h.transactionHandler.GetDashboard,
			Middleware: authMiddleware,
		},
		{
			Method:     http.MethodGet,
			Path:       "/finance/categories",
			Handler:    h.transactionHandler.GetCategories,
			Middleware: authMiddleware,
		},
		{
			Method:     http.MethodPost,
			Path:       "/finance/transactions",
			Handler:    h.transactionHandler.CreateTransaction,
			Middleware: authMiddleware,
		},
		{
			Method:     http.MethodGet,
			Path:       "/finance/transactions",
			Handler:    h.transactionHandler.GetTransactions,
			Middleware: authMiddleware,
		},
		{
			Method:     http.MethodGet,
			Path:       "/finance/transactions/{id}",
			Handler:    h.transactionHandler.GetTransaction,
			Middleware: authMiddleware,
		},
		{
			Method:     http.MethodPut,
			Path:       "/finance/transactions/{id}",
			Handler:    h.transactionHandler.UpdateTransaction,
			Middleware: authMiddleware,
		},
		{
			Method:     http.MethodDelete,
			Path:       "/finance/transactions/{id}",
			Handler:    h.transactionHandler.DeleteTransaction,
			Middleware: authMiddleware,
		},
	}
}
