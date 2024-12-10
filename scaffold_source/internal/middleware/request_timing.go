package middleware

import (
	"fmt"
	"time"

	"github.com/PROJECT_NAME/internal/logger"
	"github.com/gofiber/fiber/v2"
)

type (
	requestTimingMiddlewareDependencies interface {
		logger.LoggerProvider
	}

	RequestTimingMiddleware struct {
		d requestTimingMiddlewareDependencies
	}
)

func NewRequestTimingMiddleware(d requestTimingMiddlewareDependencies) *RequestTimingMiddleware {
	return &RequestTimingMiddleware{d: d}
}

func (m *RequestTimingMiddleware) Handle(c *fiber.Ctx) error {
	t := time.Now()

	defer func() {
		m.d.Logger().Info(fmt.Sprintf("Request path: %s method: %s took: %dms", c.Path(), c.Method(), time.Since(t).Milliseconds()))
	}()

	return c.Next()
}
