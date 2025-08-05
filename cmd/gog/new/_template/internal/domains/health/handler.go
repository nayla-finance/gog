package health

import (
	"github.com/PROJECT_NAME/internal/config"
	"github.com/PROJECT_NAME/internal/logger"
	"github.com/gofiber/fiber/v2"
)

type (
	healthHandlerDependencies interface {
		logger.LoggerProvider
		ServiceProvider
		config.ConfigProvider
	}

	// Handler handles health check requests
	handler struct {
		d healthHandlerDependencies
	}

	// HealthResponse represents health check response
	HealthResponse struct {
		Status  string `json:"status" example:"ok"`
		Message string `json:"message,omitempty" example:""`
	}
)

func NewHandler(d healthHandlerDependencies) *handler {
	return &handler{d: d}
}

func (h *handler) RegisterRoutes(r fiber.Router) {
	r.Get("/ping", h.Ping)
	r.Get("/healthz/alive", h.LivenessCheck)
	r.Get("/healthz/ready", h.ReadinessCheck)
}

// @Summary      Liveness check
// @Description  Check if the application is running
// @Tags         health
// @Accept       json
// @Produce      json
// @Success      200  {object}  HealthResponse
// @Failure      500  {object}  HealthResponse
// @Router       /healthz/alive [get]
func (h *handler) LivenessCheck(c *fiber.Ctx) error {
	if err := h.d.HealthService().LivenessCheck(c.UserContext()); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(HealthResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.JSON(HealthResponse{
		Status:  "ok",
		Message: "live and kicking! 🦁",
	})
}

// @Summary      Readiness check
// @Description  Check if the application is ready
// @Tags         health
// @Accept       json
// @Produce      json
// @Success      200  {object}  HealthResponse
// @Failure      500  {object}  HealthResponse
// @Router       /healthz/ready [get]
func (h *handler) ReadinessCheck(c *fiber.Ctx) error {
	if err := h.d.HealthService().ReadinessCheck(c.UserContext()); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(HealthResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	return c.JSON(HealthResponse{
		Status:  "ok",
		Message: "ready like a lion, ready to pounce on any incoming requests! 🦁",
	})
}

// @Summary      Ping
// @Description  Tests connectivity by pinging the application, requires authentication to verify caller identity
// @Tags         health
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  HealthResponse
// @Failure      500  {object}  HealthResponse
// @Router       /ping [get]
func (h *handler) Ping(c *fiber.Ctx) error {
	// Ping used to ping this service using its API key to make sure the connection is working
	return c.JSON(HealthResponse{
		Status: "ok",
	})
}
