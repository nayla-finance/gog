package errors

import (
	"fmt"
	"time"

	"github.com/PROJECT_NAME/internal/logger"
	"github.com/getsentry/sentry-go"
	"github.com/gofiber/fiber/v2"
)

type (
	ErrorHandlerProvider interface {
		ErrorHandler() *Handler
	}

	handlerDependencies interface {
		logger.LoggerProvider
	}

	Handler struct {
		d handlerDependencies
	}

	ErrorResponse struct {
		StatusCode int       `json:"statusCode"`
		ErrorCode  ErrorCode `json:"errorCode"`
		Message    string    `json:"message"`
		Timestamp  string    `json:"timestamp"`
		Path       string    `json:"path"`
	}
)

func NewErrorHandler(d handlerDependencies) *Handler {
	return &Handler{
		d: d,
	}
}

func (h *Handler) Handle(c *fiber.Ctx, err error) error {
	sentry.CaptureException(err)

	return h.errorResponseJSON(c, err)
}

func (h *Handler) errorResponseJSON(ctx *fiber.Ctx, err error) error {
	er := &ErrorResponse{}

	// If it's already our custom error, use it
	if appErr, ok := err.(*AppError); ok {
		er.ErrorCode = appErr.Code
		er.Message = appErr.Message
	} else {
		// Default to internal server error
		er.ErrorCode = ErrInternal
		er.Message = err.Error()
	}

	er.Path = ctx.Path()
	er.Timestamp = time.Now().Format(time.RFC3339)
	er.StatusCode = er.HttpStatus()

	h.d.Logger().Error(fmt.Sprintf("🔥🔥🔥 Error - path: %s request_id: %s status: %d error: %s",
		ctx.Path(),
		ctx.Locals("RequestID"),
		er.StatusCode,
		err.Error(),
	))

	return ctx.Status(er.StatusCode).JSON(er)
}

// Map error codes to HTTP status codes
func (e *ErrorResponse) HttpStatus() int {
	switch e.ErrorCode {
	case ErrUnauthorized:
		return fiber.StatusUnauthorized
	case ErrForbidden:
		return fiber.StatusForbidden
	case ErrBadRequest, ErrAccountAlreadyExists, ErrDuplicateEntry, ErrInvalidInput, ErrMissingField:
		return fiber.StatusBadRequest
	case ErrResourceNotFound:
		return fiber.StatusNotFound
	default:
		return fiber.StatusInternalServerError
	}
}
