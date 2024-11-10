package errors

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/project-name/internal/logger"
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

	return ctx.Status(er.StatusCode).JSON(er)
}

// Map error codes to HTTP status codes
func (e *ErrorResponse) HttpStatus() int {
	switch {
	case e.ErrorCode == ErrUnauthorized:
		return fiber.StatusUnauthorized
	case e.ErrorCode == ErrForbidden:
		return fiber.StatusForbidden
	case e.ErrorCode >= 2000 && e.ErrorCode < 3000:
		return fiber.StatusUnauthorized
	case e.ErrorCode >= 3000 && e.ErrorCode < 4000:
		return fiber.StatusBadRequest
	case e.ErrorCode >= 4000 && e.ErrorCode < 5000:
		return fiber.StatusNotFound
	default:
		return fiber.StatusInternalServerError
	}
}
