package post

import (
	"github.com/PROJECT_NAME/internal/config"
	"github.com/PROJECT_NAME/internal/errors"
	"github.com/gofiber/fiber/v2"
	"github.com/nayla-finance/go-nayla/logger"
)

// specific middlewares for post domain

type (
	specificPostMiddlewareDependencies interface {
		config.ConfigProvider
		logger.Provider
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
