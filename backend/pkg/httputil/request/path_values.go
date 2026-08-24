package request

import (
	errors_core "backend/pkg/errors"
	"fmt"
	"net/http"
	"strconv"
)

func GetIntPathValue(r *http.Request, key string) (int, error) {
	pathValue := r.PathValue(key)
	if pathValue == "" {
		return 0, fmt.Errorf("no path value for key '%s',err: %w", key, errors_core.ErrInvalidArgument)
	}

	val, err := strconv.Atoi(pathValue)
	if err != nil {
		return 0, fmt.Errorf("path value='%s' by key='%s',not a valid integer:%v:%w", pathValue, key, err, errors_core.ErrInvalidArgument)
	}
	return val, nil
}
