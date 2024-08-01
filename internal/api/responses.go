package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func BadRequestError(err error) *echo.HTTPError {
	return &echo.HTTPError{Code: http.StatusBadRequest, Message: err, Internal: nil}
}

func NotFoundError(err error) *echo.HTTPError {
	return &echo.HTTPError{Code: http.StatusNotFound, Message: err, Internal: nil}
}

func UnprocessableEntityError(err error) *echo.HTTPError {
	return &echo.HTTPError{Code: http.StatusUnprocessableEntity, Message: err, Internal: nil}
}

func TooManyRequestsError(err error) *echo.HTTPError {
	return &echo.HTTPError{Code: http.StatusTooManyRequests, Message: err, Internal: nil}
}

func InternalServerError(err error, message ...string) *echo.HTTPError {
	herr := &echo.HTTPError{Code: http.StatusInternalServerError, Message: "InternalServerError", Internal: err}
	if len(message) > 0 {
		herr.Message = message[0]
	}

	return herr
}
