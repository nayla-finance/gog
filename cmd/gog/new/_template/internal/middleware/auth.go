package middleware

import (
	"strings"

	"github.com/PROJECT_NAME/internal/config"
	"github.com/PROJECT_NAME/internal/errors"
	"github.com/PROJECT_NAME/internal/logger"
	"github.com/PROJECT_NAME/internal/utils"
	"github.com/gofiber/fiber/v2"
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
		m.d.Logger().Debug("public route skipping auth middleware")
		return c.Next()
	}

	authorization := strings.Split(c.Get("Authorization"), " ")
	if len(authorization) != 2 || authorization[0] != "Bearer" {
		m.d.Logger().Error("missing or invalid API key in Authorization header, Got token: ", authorization)
		return m.d.NewError(errors.ErrUnauthorized, "Authorization header must be in format 'Bearer <token>'")
	}

	token := authorization[1]

	if token != m.d.Config().Api.Key {
		m.d.Logger().Error("missing or invalid API key in Authorization header, Got token: ", authorization)
		return m.d.NewError(errors.ErrUnauthorized, "Missing or invalid API key in Authorization header")
	}

	return c.Next()
}

func (m *AuthMiddleware) isPublicRoute(path string) bool {
	path = utils.NormalizePath(path)
	for _, p := range m.d.Config().Api.PublicRoutes {
		if path == utils.NormalizePath(p) {
			return true
		}
	}

	return false
}
