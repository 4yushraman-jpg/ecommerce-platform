package errors

import "errors"

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenRevoked       = errors.New("token revoked")
	ErrTokenNotFound      = errors.New("token not found")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrBadRequest         = errors.New("bad request")
)
