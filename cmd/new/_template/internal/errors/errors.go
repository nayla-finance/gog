package errors

type (
	ErrorProvider interface {
		NewError(c ErrorCode, m string) *AppError
	}

	AppError struct {
		Code    ErrorCode
		Message string
	}
)

func (e *AppError) Error() string {
	return e.Message
}
