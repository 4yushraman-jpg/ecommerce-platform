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
	case errors.Is(err, ErrProductNotFound):
		return HTTPError{Status: http.StatusNotFound, Message: "product not found"}
	case errors.Is(err, ErrInvalidCursor):
		return HTTPError{Status: http.StatusBadRequest, Message: "invalid cursor"}
	case errors.Is(err, ErrInvalidSort):
		return HTTPError{Status: http.StatusBadRequest, Message: "invalid sort"}
	case errors.Is(err, ErrNoUpdateFields):
		return HTTPError{Status: http.StatusBadRequest, Message: "at least one field must be provided"}
	default:
		return HTTPError{Status: http.StatusInternalServerError, Message: "internal server error"}
	}
}
