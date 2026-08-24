package http_web

import (
	"backend/pkg/server/http"
)

type WebHTTPHandler struct {
	webService WebService
}

type WebService interface {
	GetMainPage() ([]byte, error)
}

func NewWebHTTPHandler(webService WebService) *WebHTTPHandler {
	return &WebHTTPHandler{
		webService: webService,
	}
}

func (h *WebHTTPHandler) Routes() []http.Route {
	return []http.Route{
		{
			Path:    "/",
			Handler: h.GetMainPage,
		},
	}
}
