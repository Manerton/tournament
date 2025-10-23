package errs

import (
	"fmt"
	"net/http"
)

type ApiError struct {
	Code     string
	HttpCode int
	Message  string
}

func (e *ApiError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *ApiError) Wrap(message string) *ApiError {
	return &ApiError{
		Code:     e.Code,
		HttpCode: e.HttpCode,
		Message:  message,
	}
}

func IsApiError(err error) (*ApiError, bool) {
	apiErr, ok := err.(*ApiError)
	return apiErr, ok
}

const (
	DublicateError = "23505"
)

// Common errors
var (
	ErrInternalError = &ApiError{Code: "INTERNAL_SERVER_ERROR", HttpCode: http.StatusInternalServerError, Message: "Something was wrong"}
	ErrBadRequest    = &ApiError{Code: "BAD_REQUEST", HttpCode: http.StatusBadRequest, Message: "Invalid data"}
	ErrNotFound      = &ApiError{Code: "RECORD_NOT_FOUND", HttpCode: http.StatusNotFound, Message: "record not found"}
	ErrInvalidId     = &ApiError{Code: "INVALID_ID", HttpCode: http.StatusBadRequest, Message: "Invalid ID"}
)

// Team errors
var (
	ErrTeamExist = &ApiError{Code: "TEAM ALREADY EXIST", HttpCode: http.StatusConflict, Message: "team already exist"}
)

// Event errors
var (
	ErrStartTournament = &ApiError{Code: "EVENT CAN`T START", HttpCode: http.StatusNoContent, Message: "teams is 0, can't start tournament"}
	ErrFinishedEvent   = &ApiError{Code: "EVENT CAN`T FINISHED", HttpCode: http.StatusConflict, Message: "event cant finished"}
)
