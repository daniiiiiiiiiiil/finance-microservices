package user_transport_http

import (
	"backend/internal/core/auth/jwt"
	"backend/internal/core/domain"
	core_http_middleware "backend/internal/core/transport/http/middleware"
	"backend/internal/core/transport/http/server"
	"context"
	"net/http"
)

type UsersHTTPHandler struct {
	usersService UsersService
	jwtManager   *jwt.JWTManager
}

type UsersService interface {
	GetUser(ctx context.Context, id int) (domain.User, error)
	DeleteUser(ctx context.Context, id int) error
	PatchUser(ctx context.Context, id int, user domain.UserPatch) (domain.User, error)
}

func NewUsersHTTPHandler(usersService UsersService, jwtManager *jwt.JWTManager) *UsersHTTPHandler {
	return &UsersHTTPHandler{
		usersService: usersService,
		jwtManager:   jwtManager,
	}
}

func (h *UsersHTTPHandler) Routes() []server.Route {
	return []server.Route{
		{
			Method:  http.MethodGet,
			Path:    "/users/{id}",
			Handler: h.GetUser,
			Middleware: []core_http_middleware.Middleware{
				core_http_middleware.Auth(h.jwtManager),
			},
		},
		{
			Method:  http.MethodDelete,
			Path:    "/users/{id}",
			Handler: h.DeleteUser,
			Middleware: []core_http_middleware.Middleware{
				core_http_middleware.Auth(h.jwtManager),
			},
		},
		{
			Method:  http.MethodPatch,
			Path:    "/users/{id}",
			Handler: h.PatchUser,
			Middleware: []core_http_middleware.Middleware{
				core_http_middleware.Auth(h.jwtManager),
			},
		},
	}
}
