package errors

import "net/http"

func NewBadRequestError(message string) ResponseError {
	return ResponseError{
		Status: http.StatusBadRequest,
		Message: message,
	}
}

func NewServerError(message string) ResponseError {
	return ResponseError{
		Status: http.StatusInternalServerError,
		Message: message,
	}
}
