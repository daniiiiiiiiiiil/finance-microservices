package pool

import "errors"

var (
	ErrNoRows             = errors.New("no rows")
	ErrViolatesForeignKey = errors.New("violates foreign key")
	ErrConflict           = errors.New("conflict")
	ErrUnknown            = errors.New("unknown error")
)
