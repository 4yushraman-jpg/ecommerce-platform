package errors

import "errors"

var (
	ErrBadRequest      = errors.New("bad request")
	ErrProductNotFound = errors.New("product not found")
	ErrInvalidCursor   = errors.New("invalid cursor")
	ErrInvalidSort     = errors.New("invalid sort")
	ErrNoUpdateFields  = errors.New("no update fields provided")
)
