package health

import (
	"github.com/gofiber/fiber/v2"
	"github.com/project-name/internal/logger"
)

type (
	healthHandlerDependencies interface {
		logger.LoggerProvider
	}

	HealthHandler struct {
		d healthHandlerDependencies
	}
)

func NewHealthHandler(d healthHandlerDependencies) *HealthHandler {
	return &HealthHandler{d: d}
}

// @Summary		Health check
// @Description	Check if the application is running
// @Tags			health
// @Accept			json
// @Produce		json
// @Success		200	{object}	map[string]string
// @Router			/health [get]
func (h *HealthHandler) HealthCheck(c *fiber.Ctx) error {
	h.d.Logger().Info("Health check")
	return c.JSON(fiber.Map{"status": "ok"})
}
