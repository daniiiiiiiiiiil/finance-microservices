package http_auth

import (
	"backend/internal/core/transport/http/server"
	"net/http"
)

func (h *AuthHandler) Routes() []server.Route {
	return []server.Route{
		{
			Method:  http.MethodPost,
			Path:    "/auth/register",
			Handler: h.Register,
		},
		{
			Method:  http.MethodPost,
			Path:    "/auth/login",
			Handler: h.Login,
		},
		{
			Method:  http.MethodPost,
			Path:    "/auth/logout",
			Handler: h.Logout,
		},
	}
}
