// Admin settings endpoints.
package settings

import (
	"github.com/gofiber/fiber/v3"

	"github.com/skycode/ojt-management/backend/internal/audit"
	"github.com/skycode/ojt-management/backend/internal/auth"
	"github.com/skycode/ojt-management/backend/internal/platform/db"
	"github.com/skycode/ojt-management/backend/internal/platform/httpx"
)

type Handler struct {
	pool *db.Pool
}

func NewHandler(pool *db.Pool) *Handler { return &Handler{pool: pool} }

// Get: GET /api/v1/admin/settings
func (h *Handler) Get(c fiber.Ctx) error {
	s, err := Get(c.Context(), h.pool)
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	return httpx.OK(c, s)
}

// Patch: PATCH /api/v1/admin/settings
func (h *Handler) Patch(c fiber.Ctx) error {
	var p Patch
	if err := c.Bind().JSON(&p); err != nil {
		return httpx.Fail(c, httpx.ValidationMsg("Malformed request body."))
	}
	actor := auth.ActorOf(c)
	s, err := Update(c.Context(), h.pool, p, actor.ID)
	if err != nil {
		return httpx.Fail(c, err)
	}
	audit.Record(c.Context(), h.pool, &actor.ID, "settings.updated", "institution_settings", &s.ID,
		httpx.RequestID(c), map[string]any{"fields": fieldNames(p)})
	return httpx.OK(c, s)
}

func fieldNames(p Patch) []string {
	var names []string
	if p.Timezone != nil {
		names = append(names, "timezone")
	}
	if p.LowAccuracyThresholdM != nil {
		names = append(names, "low_accuracy_threshold_m")
	}
	if p.NearCompletionPercent != nil {
		names = append(names, "near_completion_percent")
	}
	if p.MissingJournalCutoffHours != nil {
		names = append(names, "missing_journal_cutoff_hours")
	}
	if p.UnusualSessionMinutes != nil {
		names = append(names, "unusual_session_minutes")
	}
	if p.RetentionDays != nil {
		names = append(names, "retention_days")
	}
	return names
}
