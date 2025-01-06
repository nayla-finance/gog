package middleware

import (
	"github.com/PROJECT_NAME/internal/errors"
	"github.com/PROJECT_NAME/internal/logger"
	"github.com/gofiber/fiber/v2"
)

type (
	notFoundMiddlewareDependencies interface {
		logger.LoggerProvider
		errors.ErrorProvider
	}

	NotFoundMiddleware struct {
		d notFoundMiddlewareDependencies
	}
)

func NewNotFoundMiddleware(d notFoundMiddlewareDependencies) *NotFoundMiddleware {
	return &NotFoundMiddleware{d: d}
}

func (m *NotFoundMiddleware) Handle(c *fiber.Ctx) error {
	m.d.Logger().Error("Not found", "path", c.Path())
	return m.d.NewError(errors.ErrResourceNotFound, "route not found")
}
