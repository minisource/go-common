package health

import (
	"github.com/gofiber/fiber/v2"
)

// FiberHandler provides HTTP handlers for health checks
type FiberHandler struct {
	healthService *HealthService
}

// NewFiberHandler creates a new Fiber health handler
func NewFiberHandler(healthService *HealthService) *FiberHandler {
	return &FiberHandler{
		healthService: healthService,
	}
}

// Liveness handles liveness probe requests
// @Summary Liveness check
// @Description Check if service is alive
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (h *FiberHandler) Liveness(c *fiber.Ctx) error {
	result := h.healthService.CheckLiveness()
	return c.JSON(result)
}

// Readiness handles readiness probe requests
// @Summary Readiness check
// @Description Check if service is ready to accept traffic
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 503 {object} map[string]interface{}
// @Router /ready [get]
func (h *FiberHandler) Readiness(c *fiber.Ctx) error {
	result, healthy := h.healthService.CheckReadiness(c.Context())

	if !healthy {
		return c.Status(fiber.StatusServiceUnavailable).JSON(result)
	}

	return c.JSON(result)
}

// RegisterRoutes registers health check routes
func (h *FiberHandler) RegisterRoutes(app *fiber.App) {
	app.Get("/health", h.Liveness)
	app.Get("/ready", h.Readiness)
	app.Get("/healthz", h.Liveness) // Kubernetes standard
	app.Get("/readyz", h.Readiness) // Kubernetes standard
}
