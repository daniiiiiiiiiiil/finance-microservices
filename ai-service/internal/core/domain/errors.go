package domain

import "errors"

var (
	ErrInvalidInput        = errors.New("invalid input")
	ErrAddressNotFound     = errors.New("address not found")
	ErrNoStoresFound       = errors.New("no stores found")
	ErrProviderUnavailable = errors.New("provider unavailable")
	ErrProviderTimeout     = errors.New("provider timeout")
	ErrProviderBadResponse = errors.New("provider bad response")
)
