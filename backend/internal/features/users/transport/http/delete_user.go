package user_transport_http

import (
	"backend/internal/core/logger"
	core_http_middleware "backend/internal/core/transport/http/middleware"
	"backend/internal/core/transport/http/request"
	"backend/internal/core/transport/http/response"
	"errors"
	"net/http"
)

// DeleteUser godoc
// @Summary Удаление пользователя
// @Description Удаление существующего пользователя по его ID
// @Tags users
// @Param id path int true "ID удаляемого пользователя"
// @Success 204 "Успешное удаление пользователя"
// @Failure 400 {object} response_core.ErrorResponse "Bad request"
// @Failure 404 {object} response_core.ErrorResponse "User not found"
// @Failure 500 {object} response_core.ErrorResponse "Internal server error"
// @Router /users/{id} [delete]
func (h *UsersHTTPHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := response_core.NewHTTPResponseHandler(log, w)
	id, err := request.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get userID path value")
		return
	}

	currentUserID, ok := core_http_middleware.GetUserID(r)
	if !ok {
		responseHandler.ErrorResponse(errors.New("unauthorized"), "user not authenticated")
		return
	}
	if id != currentUserID {
		responseHandler.ErrorResponse(errors.New("forbidden"), "you can only delete your own account")
		return
	}
	if err := h.usersService.DeleteUser(ctx, id); err != nil {
		responseHandler.ErrorResponse(err, "failed to delete user")
		return
	}
	responseHandler.NoContentResponse()
}
