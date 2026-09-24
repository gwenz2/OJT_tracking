// HTTP handlers for /api/v1/staff/reports/{name}[/export.csv].
package reports

import (
	"bytes"

	"github.com/gofiber/fiber/v3"

	"github.com/skycode/ojt-management/backend/internal/auth"
	"github.com/skycode/ojt-management/backend/internal/platform/db"
	"github.com/skycode/ojt-management/backend/internal/platform/httpx"
)

type Handler struct {
	pool *db.Pool
}

func NewHandler(pool *db.Pool) *Handler { return &Handler{pool: pool} }

func filtersOf(c fiber.Ctx) Filters {
	return Filters{
		From: c.Query("from"), To: c.Query("to"),
		TraineeID: c.Query("trainee_id"), SiteID: c.Query("site_id"),
		Status: c.Query("status"),
	}
}

// Get: GET /staff/reports/{name} — JSON rows + columns.
func (h *Handler) Get(c fiber.Ctx) error {
	name := c.Params("name")
	scope, err := auth.ScopedTraineeIDs(c.Context(), h.pool, auth.ActorOf(c))
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	r, err := Run(c.Context(), h.pool, name, scope, filtersOf(c))
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, r)
}

// Export: GET /staff/reports/{name}/export.csv — CSV download.
func (h *Handler) Export(c fiber.Ctx) error {
	name := c.Params("name")
	scope, err := auth.ScopedTraineeIDs(c.Context(), h.pool, auth.ActorOf(c))
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	f := filtersOf(c)
	r, err := Run(c.Context(), h.pool, name, scope, f)
	if err != nil {
		return httpx.Fail(c, err)
	}
	var buf bytes.Buffer
	if err := WriteCSV(&buf, r); err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", `attachment; filename="`+Filename(name, f)+`"`)
	return c.Send(buf.Bytes())
}
