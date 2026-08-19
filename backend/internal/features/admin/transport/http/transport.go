package http

import (
	"backend/internal/core/auth/jwt"
	core_middleware "backend/internal/core/transport/http/middleware"
	"backend/internal/core/transport/http/server"
	"backend/internal/features/admin/service"
	"net/http"
)

type AdminHandler struct {
	service    *service_admin.AdminService
	jwtManager *jwt.JWTManager
}

func NewAdminHandler(s *service_admin.AdminService, jwtManager *jwt.JWTManager) *AdminHandler {
	return &AdminHandler{
		service:    s,
		jwtManager: jwtManager,
	}
}

func (h *AdminHandler) isAdmin(r *http.Request) bool {
	return core_middleware.IsAdmin(r)
}

func (h *AdminHandler) Routes() []server.Route {
	authMiddleware := []core_middleware.Middleware{
		core_middleware.Auth(h.jwtManager),
	}

	return []server.Route{
		{
			Method:     http.MethodGet,
			Path:       "/admin/users",
			Handler:    h.GetUsers,
			Middleware: authMiddleware,
		},
		{
			Method:     http.MethodGet,
			Path:       "/admin/users/{id}",
			Handler:    h.GetUser,
			Middleware: authMiddleware,
		},
		{
			Method:     http.MethodDelete,
			Path:       "/admin/users/{id}",
			Handler:    h.DeleteUser,
			Middleware: authMiddleware,
		},
		{
			Method:     http.MethodPatch,
			Path:       "/admin/users/{id}/role",
			Handler:    h.UpdateUserRole,
			Middleware: authMiddleware,
		},
		{
			Method:     http.MethodGet,
			Path:       "/admin/metrics",
			Handler:    h.GetMetrics,
			Middleware: authMiddleware,
		},
	}
}
