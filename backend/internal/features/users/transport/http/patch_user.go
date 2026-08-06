package user_transport_http

import (
	"backend/internal/core/domain"
	"backend/internal/core/logger"
	core_http_middleware "backend/internal/core/transport/http/middleware"
	request2 "backend/internal/core/transport/http/request"
	"backend/internal/core/transport/http/response"
	"backend/internal/core/transport/http/types"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type PatchUserRequest struct {
	FullName    types.Nullable[string] `json:"full_name" swaggertype:"string" example:"John Doe"`
	PhoneNumber types.Nullable[string] `json:"phone_number" swaggertype:"string" example:"+799999999"`
}

func (r *PatchUserRequest) Validate() error {
	if r.FullName.Set {
		if r.FullName.Value == nil {
			return fmt.Errorf("full name is required")
		}
		fullName := len([]rune(*r.FullName.Value))
		if fullName < 3 || fullName > 100 {
			return fmt.Errorf("full name must be between 3 and 100 characters")
		}
	}
	if r.PhoneNumber.Set {
		if r.PhoneNumber.Value == nil {
			return fmt.Errorf("phone number is required")
		}
		phoneNumber := len([]rune(*r.PhoneNumber.Value))
		if phoneNumber < 10 || phoneNumber > 15 {
			return fmt.Errorf("phone number must be between 10 and 15 characters")
		}
		if !strings.HasPrefix(*r.PhoneNumber.Value, "+") {
			return fmt.Errorf("phone number must start with '+'")
		}
	}
	return nil
}

type PatchUserResponse UserDTOResponse

// PatchUser godoc
// @Summary Изменение пользователя
// @Description Изменения информации о существующем пользователе по его ID
// @Description ### Логика обновления полей
// @Description 1 Поле не передано `phone_number` игнорируется, значение в бд не меняется
// @Description 2 Явно передано значения `"phone_number":"+799999999"` - устанавливает новый номер телефона
// @Description 3 Передан null `"phone_number": null` - очищает поле в БД (set to NULL)
// @Description Ограничение: `full_name` не может быть выставлен как null
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "ID изменяемого пользователя"
// @Param request body PatchUserRequest true "PatchUser тело запроса"
// @Success 200 {object} PatchUserResponse "Успешно изменились данные о пользователе"
// @Failure 400 {object} response_core.ErrorResponse "Bad request"
// @Failure 404 {object} response_core.ErrorResponse "User not found"
// @Failure 409 {object} response_core.ErrorResponse "Conflict"
// @Failure 500 {object} response_core.ErrorResponse "Internal server error"
// @Router /users/{id} [patch]
func (h *UsersHTTPHandler) PatchUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	responseHandler := response_core.NewHTTPResponseHandler(log, w)

	userID, err := request2.GetIntPathValue(r, "id")
	if err != nil {
		responseHandler.ErrorResponse(err, "failed get ID is required")
		return
	}

	currentUserID, ok := core_http_middleware.GetUserID(r)
	if !ok {
		responseHandler.ErrorResponse(errors.New("unauthorized"), "user not authenticated")
		return
	}
	if userID != currentUserID {
		responseHandler.ErrorResponse(errors.New("forbidden"), "you can only update your own profile")
		return
	}

	var request PatchUserRequest
	if err := request2.DecodeAndValidateRequest(r, &request); err != nil {
		responseHandler.ErrorResponse(err, "failed to decode and validate request")
		return
	}

	userPatch := userPatchFromRequest(request)
	userDomain, err := h.usersService.PatchUser(ctx, userID, userPatch)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to patch user")
		return
	}

	response := PatchUserResponse(UserDTOFromDomain(userDomain))
	responseHandler.JSONResponse(response, http.StatusOK)
}

func userPatchFromRequest(request PatchUserRequest) domain.UserPatch {
	return domain.NewUserPatch(
		request.FullName.ToDomain(),
		request.PhoneNumber.ToDomain())
}
