package errors

import (
	"errors"
	"net/http"
)

type HTTPError struct {
	Status  int
	Message string
}

func ToHTTPError(err error) HTTPError {
	switch {
	case errors.Is(err, ErrBadRequest):
		return HTTPError{Status: http.StatusBadRequest, Message: "bad request"}
	case errors.Is(err, ErrUserAlreadyExists):
		return HTTPError{Status: http.StatusConflict, Message: "user already exists"}
	case errors.Is(err, ErrInvalidCredentials):
		return HTTPError{Status: http.StatusUnauthorized, Message: "invalid credentials"}
	case errors.Is(err, ErrInvalidToken):
		return HTTPError{Status: http.StatusUnauthorized, Message: "invalid token"}
	case errors.Is(err, ErrTokenExpired):
		return HTTPError{Status: http.StatusUnauthorized, Message: "token expired"}
	case errors.Is(err, ErrTokenRevoked):
		return HTTPError{Status: http.StatusUnauthorized, Message: "token revoked"}
	case errors.Is(err, ErrTokenNotFound):
		return HTTPError{Status: http.StatusUnauthorized, Message: "token not found"}
	case errors.Is(err, ErrUnauthorized):
		return HTTPError{Status: http.StatusUnauthorized, Message: "unauthorized"}
	default:
		return HTTPError{Status: http.StatusInternalServerError, Message: "internal server error"}
	}
}
