package server

import (
	core_http_middleware "backend/internal/core/transport/http/middleware"
	"net/http"
)

type Route struct {
	Method     string
	Path       string
	Handler    http.HandlerFunc
	Middleware []core_http_middleware.Middleware
}

func (r *Route) WithMiddleware() http.Handler {
	return core_http_middleware.ChainMiddlewares(r.Handler, r.Middleware...)
}
