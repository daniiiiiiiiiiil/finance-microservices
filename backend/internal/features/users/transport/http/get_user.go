package user_transport_http

import (
	"backend/internal/core/logger"
	core_http_middleware "backend/internal/core/transport/http/middleware"
	"backend/internal/core/transport/http/request"
	"backend/internal/core/transport/http/response"
	"errors"
	"net/http"
)

type GetUserResponse UserDTOResponse

// GetUser godoc
// @Summary Получение пользователя
// @Description Получение существующего пользователя по его ID
// @Tags users
// @Produce json
// @Param id path int true "ID получаемого пользователя"
// @Success 200 {object} GetUserResponse "Успешно найден пользователь"
// @Failure 400 {object} response_core.ErrorResponse "Bad request"
// @Failure 404 {object} response_core.ErrorResponse "User not found"
// @Failure 500 {object} response_core.ErrorResponse "Internal server error"
// @Router /users/{id} [get]
func (h *UsersHTTPHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := response_core.NewHTTPResponseHandler(log, w)
	userID, err := request.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get user id from path")
		return
	}
	currentUserID, ok := core_http_middleware.GetUserID(r)
	if !ok {
		responseHandler.ErrorResponse(errors.New("unauthorized"), "user not authenticated")
		return
	}
	if userID != currentUserID {
		responseHandler.ErrorResponse(errors.New("forbidden"), "you can only view your own profile")
		return
	}
	user, err := h.usersService.GetUser(ctx, userID)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get user")
		return
	}
	userDTO := GetUserResponse(UserDTOFromDomain(user))

	responseHandler.JSONResponse(userDTO, http.StatusOK)

}
