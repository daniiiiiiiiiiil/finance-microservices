package errors

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrInvalidArgument   = errors.New("invalid argument")
	ErrConflict          = errors.New("conflict")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrUnavailable       = errors.New("unavailable")
	ErrUnexpected        = errors.New("unexpected response")
	ErrTelegramForbidden = errors.New("telegram forbidden")
)
