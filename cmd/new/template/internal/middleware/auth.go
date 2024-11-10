package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/project-name/internal/config"
	"github.com/project-name/internal/errors"
	"github.com/project-name/internal/logger"
)

type (
	authMiddlewareDependencies interface {
		config.ConfigProvider
		logger.LoggerProvider
		errors.ErrorProvider
	}

	AuthMiddleware struct {
		d authMiddlewareDependencies
	}
)

func NewAuthMiddleware(d authMiddlewareDependencies) *AuthMiddleware {
	return &AuthMiddleware{d: d}
}

func (m *AuthMiddleware) Handle(c *fiber.Ctx) error {
	if m.isPublicRoute(c.Path()) {
		m.d.Logger().Info("public route skipping auth middleware")
		return c.Next()
	}

	if c.Get("X-API-KEY") != m.d.Config().ApiKey {
		return m.d.NewError(errors.ErrUnauthorized, "Missing or invalid API key")
	}

	return c.Next()
}

func (m *AuthMiddleware) isPublicRoute(path string) bool {
	publicPaths := []string{
		"/api/health",
		"/api/docs",
	}

	for _, p := range publicPaths {
		if strings.HasPrefix(path, p) {
			return true
		}
	}

	return false
}
