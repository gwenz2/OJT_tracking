// Health endpoints: liveness (process up) and readiness (essential
// dependencies reachable). Readiness reports no sensitive detail.
package httpx

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
)

type HealthDeps struct {
	CheckDB      func(ctx context.Context) error
	CheckStorage func(ctx context.Context) error
}

func RegisterHealth(app *fiber.App, deps HealthDeps) {
	app.Get("/health/live", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	app.Get("/health/ready", func(c fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 4*time.Second)
		defer cancel()

		checks := fiber.Map{}
		ready := true
		if err := deps.CheckDB(ctx); err != nil {
			checks["database"] = "down"
			ready = false
		} else {
			checks["database"] = "ok"
		}
		if deps.CheckStorage != nil {
			if err := deps.CheckStorage(ctx); err != nil {
				checks["storage"] = "down"
				ready = false
			} else {
				checks["storage"] = "ok"
			}
		}
		status := fiber.StatusOK
		if !ready {
			status = fiber.StatusServiceUnavailable
		}
		return c.Status(status).JSON(fiber.Map{"ready": ready, "checks": checks})
	})
}
