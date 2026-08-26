package response

import (
	errors_core "github.com/daniiiiiiiiiiil/finance-microservices/gateway-service/pkg/errors"
	"github.com/daniiiiiiiiiiil/finance-microservices/gateway-service/pkg/logger"

	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"go.uber.org/zap"
)

type HTTPResponseHandler struct {
	log *logger.Logger
	rw  http.ResponseWriter
}

func NewHTTPResponseHandler(log *logger.Logger, rw http.ResponseWriter) *HTTPResponseHandler {
	return &HTTPResponseHandler{
		log: log,
		rw:  rw,
	}
}

func (h *HTTPResponseHandler) PanicResponse(p any, msg string) {
	statusCode := http.StatusInternalServerError
	err := fmt.Errorf("unexpected panic:%v", msg)
	h.log.Error(msg,
		zap.Error(err),
		zap.Any("panic_value", p),
		zap.String("stack", string(debug.Stack())),
	)
	h.errorResponse(statusCode, err, msg)
}

func (h *HTTPResponseHandler) NoContentResponse() {
	h.rw.WriteHeader(http.StatusNoContent)
}

func (h *HTTPResponseHandler) HTMLResponse(html []byte) {
	h.rw.WriteHeader(http.StatusOK)
	h.rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := h.rw.Write(html); err != nil {
		h.log.Error("failed to write HTML response", zap.Error(err))
	}
}

func (h *HTTPResponseHandler) ErrorResponse(err error, msg string) {
	var (
		statusCode int
		logFubc    func(string, ...zap.Field)
	)
	switch {
	case errors.Is(err, errors_core.ErrInvalidArgument):
		statusCode = http.StatusBadRequest
		logFubc = h.log.Warn

	case errors.Is(err, errors_core.ErrNotFound):
		statusCode = http.StatusNotFound
		logFubc = h.log.Debug

	case errors.Is(err, errors_core.ErrConflict):
		statusCode = http.StatusConflict
		logFubc = h.log.Warn

	default:
		statusCode = http.StatusInternalServerError
		logFubc = h.log.Error
	}
	logFubc(msg, zap.Error(err))

	h.errorResponse(statusCode, err, msg)

}

func (h *HTTPResponseHandler) errorResponse(statusCode int, err error, msg string) {
	h.rw.WriteHeader(statusCode)

	response := ErrorResponse{
		Error:   err.Error(),
		Message: msg,
	}
	h.JSONResponse(response, statusCode)
}

func (h *HTTPResponseHandler) JSONResponse(responseBody any, statusCode int) {
	if err := json.NewEncoder(h.rw).Encode(responseBody); err != nil {
		h.log.Error("Write HTTP response failed", zap.Error(err))
	}
}
