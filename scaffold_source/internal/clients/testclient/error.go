package testclient

import "github.com/PROJECT_NAME/internal/errors"

type ErrorCode int

const (
	ErrResourceNotFound ErrorCode = 4000
)

type ErrorResponse struct {
	StatusCode int       `json:"statusCode"`
	ErrorCode  ErrorCode `json:"errorCode"`
	Message    string    `json:"message"`
	Timestamp  string    `json:"timestamp"`
	Path       string    `json:"path"`
}

func (e *ErrorResponse) mapError() *errors.AppError {
	switch e.ErrorCode {
	case ErrResourceNotFound:
		return &errors.AppError{
			Code:    errors.ErrResourceNotFound,
			Message: e.Message,
		}
	default:
		return &errors.AppError{
			Code:    errors.ErrInternal,
			Message: e.Message,
		}
	}
}
