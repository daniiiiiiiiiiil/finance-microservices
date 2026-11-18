package service

import (
	"errors"
)

var (
	ErrNotLinked           = errors.New("telegram account is not linked")
	ErrAlreadyLinked       = errors.New("telegram account is already linked")
	ErrCodeInvalid         = errors.New("link code is invalid or expired")
	ErrCodeUsed            = errors.New("link code has already been used")
	ErrNoActiveDialog      = errors.New("no active dialog")
	ErrInvalidInput        = errors.New("invalid input")
	ErrAccessDenied        = errors.New("access denied")
	ErrTelegramUnavailable = errors.New("telegram api is unavailable")
)
