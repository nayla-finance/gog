package health

import (
	"github.com/PROJECT_NAME/internal/db"
	"github.com/PROJECT_NAME/internal/logger"
	"github.com/gofiber/fiber/v2"
)

type (
	healthHandlerDependencies interface {
		logger.LoggerProvider
		db.DBProvider
	}

	HealthHandler struct {
		d healthHandlerDependencies
	}

	HealthCheckResponse struct {
		Status  string `json:"status"`
		Message string `json:"message,omitempty"`
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
// @Success		200	{object}	HealthCheckResponse
// @Failure		500	{object}	HealthCheckResponse
// @Router			/health [get]
func (h *HealthHandler) HealthCheck(c *fiber.Ctx) error {
	h.d.Logger().Info("Health check")

	if err := h.d.DB().Ping(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(HealthCheckResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.JSON(HealthCheckResponse{
		Status: "ok",
	})
}
