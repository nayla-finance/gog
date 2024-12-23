package middleware

import (
	"fmt"
	"time"

	"github.com/PROJECT_NAME/internal/logger"
	"github.com/gofiber/fiber/v2"
)

type (
	loggingMiddlewareDependencies interface {
		logger.LoggerProvider
	}

	LoggingMiddleware struct {
		d loggingMiddlewareDependencies
	}
)

func NewLoggingMiddleware(d loggingMiddlewareDependencies) *LoggingMiddleware {
	return &LoggingMiddleware{d: d}
}

func (m *LoggingMiddleware) Handle(c *fiber.Ctx) error {
	// Log request details before processing
	m.d.Logger().Info(fmt.Sprintf("📥 Incoming request - path: %s query: %s method: %s request_id: %s ip: %s user_agent: %s content_length: %d",
		c.Path(),
		string(c.Request().URI().QueryString()),
		c.Method(),
		c.Locals("RequestID"),
		c.IP(),
		string(c.Request().Header.UserAgent()),
		int64(len(c.Body())),
	))
	start := time.Now()

	// Process request
	err := c.Next()

	// Log response details after processing
	m.d.Logger().Info(fmt.Sprintf("📤 Outgoing response - path: %s request_id: %s duration: %dms",
		c.Path(),
		c.Locals("RequestID"),
		time.Since(start).Milliseconds(),
	))

	c.Response().Header.Set("X-Response-Time-In-Millis", fmt.Sprintf("%dms", time.Since(start).Milliseconds()))

	return err
}
