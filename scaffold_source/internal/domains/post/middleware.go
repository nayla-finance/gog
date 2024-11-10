package post

import (
	"github.com/gofiber/fiber/v2"
	"github.com/project-name/internal/config"
	"github.com/project-name/internal/errors"
	"github.com/project-name/internal/logger"
)

// specific middlewares for post domain

type (
	specificPostMiddlewareDependencies interface {
		config.ConfigProvider
		logger.LoggerProvider
		errors.ErrorProvider
	}

	SpecificPostMiddleware struct {
		d specificPostMiddlewareDependencies
	}
)

func NewSpecificPostMiddleware(d specificPostMiddlewareDependencies) *SpecificPostMiddleware {
	return &SpecificPostMiddleware{d: d}
}

func (h *SpecificPostMiddleware) Handle(c *fiber.Ctx) error {

	// do something special here ...

	return c.Next()
}
