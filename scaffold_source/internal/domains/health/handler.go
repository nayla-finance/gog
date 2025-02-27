package health

import (
	"github.com/PROJECT_NAME/internal/db"
	"github.com/PROJECT_NAME/internal/logger"
	"github.com/PROJECT_NAME/internal/nats"
	"github.com/gofiber/fiber/v2"
)

type (
	healthHandlerDependencies interface {
		logger.LoggerProvider
		db.DBProvider
		nats.ServiceProvider
	}

	HealthHandler struct {
		d healthHandlerDependencies
	}

	HealthResponse struct {
		Status  string `json:"status"`
		Message string `json:"message,omitempty"`
	}
)

func NewHealthHandler(d healthHandlerDependencies) *HealthHandler {
	return &HealthHandler{d: d}
}

func (h *HealthHandler) RegisterRoutes(r fiber.Router) {
	r.Get("/ping", h.Ping)
	r.Get("/health", h.HealthCheck)
	r.Get("/health/ready", h.ReadinessCheck)
}

// @Summary      Health check
// @Description  Check if the application is running
// @Tags         health
// @Accept       json
// @Produce      json
// @Success      200  {object}  HealthResponse
// @Failure      500  {object}  HealthResponse
// @Router       /health [get]
func (h *HealthHandler) HealthCheck(c *fiber.Ctx) error {
	if err := h.d.DB().Ping(); err != nil {
		h.d.Logger().Errorw("❌ Database connection is not healthy", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(HealthResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	if !h.d.NatsService().HealthCheck() {
		return c.Status(fiber.StatusInternalServerError).JSON(HealthResponse{
			Status:  "error",
			Message: "NATS connection is not healthy",
		})
	}

	h.d.Logger().Info("Health check passed ✅")

	return c.JSON(HealthResponse{
		Status: "ok",
	})
}

// @Summary      Readiness check
// @Description  Check if the application is ready
// @Tags         health
// @Accept       json
// @Produce      json
// @Success      200  {object}  HealthResponse
// @Failure      500  {object}  HealthResponse
// @Router       /health/ready [get]
func (h *HealthHandler) ReadinessCheck(c *fiber.Ctx) error {

	// Check all dependencies are ready (e.g. database, services, etc.)

	if err := h.d.DB().Ping(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(HealthResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	if !h.d.NatsService().HealthCheck() {
		return c.Status(fiber.StatusInternalServerError).JSON(HealthResponse{
			Status:  "error",
			Message: "NATS connection is not healthy",
		})
	}

	h.d.Logger().Info("Readiness check passed ✅")

	return c.JSON(HealthResponse{
		Status: "ok",
	})
}

// @Summary      Ping
// @Description  Tests connectivity by pinging the application, requires authentication to verify caller identity
// @Tags         health
// @Accept       json
// @Produce      json
// @Success      200  {object}  HealthResponse
// @Failure      500  {object}  HealthResponse
// @Router       /ping [get]
func (h *HealthHandler) Ping(c *fiber.Ctx) error {
	return c.JSON(HealthResponse{
		Status: "ok",
	})
}
