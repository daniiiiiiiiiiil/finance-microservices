package http

import (
	core_http_middleware "backend/pkg/middleware/http"
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
