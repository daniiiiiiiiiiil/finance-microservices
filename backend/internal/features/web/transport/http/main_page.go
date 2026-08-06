package http_web

import (
	"backend/internal/core/logger"
	response_core "backend/internal/core/transport/http/response"
	"net/http"
)

func (h *WebHTTPHandler) GetMainPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := response_core.NewHTTPResponseHandler(log, w)
	html, err := h.webService.GetMainPage()
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to fetch main page")
		return
	}
	responseHandler.HTMLResponse(html)
}
