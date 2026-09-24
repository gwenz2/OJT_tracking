// HTTP handlers for staff dashboard + attendance monitoring.
package monitoring

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/skycode/ojt-management/backend/internal/auth"
	"github.com/skycode/ojt-management/backend/internal/platform/db"
	"github.com/skycode/ojt-management/backend/internal/platform/httpx"
)

type Handler struct {
	pool *db.Pool
}

func NewHandler(pool *db.Pool) *Handler { return &Handler{pool: pool} }

// Dashboard: GET /staff/dashboard?date=
func (h *Handler) Dashboard(c fiber.Ctx) error {
	scope, err := auth.ScopedTraineeIDs(c.Context(), h.pool, auth.ActorOf(c))
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	d, err := Dashboard(c.Context(), h.pool, scope, time.Now())
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	return httpx.OK(c, d)
}

// AttendanceList: GET /staff/attendance?date=&from=&to=&site_id=&trainee_id=&status=
func (h *Handler) AttendanceList(c fiber.Ctx) error {
	scope, err := auth.ScopedTraineeIDs(c.Context(), h.pool, auth.ActorOf(c))
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	page := httpx.ParsePage(c)
	f := struct{ From, To, Date, SiteID, TraineeID, Status string }{
		From: c.Query("from"), To: c.Query("to"), Date: c.Query("date"),
		SiteID: c.Query("site_id"), TraineeID: c.Query("trainee_id"),
		Status: c.Query("status"),
	}
	items, total, err := AttendanceList(c.Context(), h.pool, scope, f, page)
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	return httpx.OKMeta(c, items, httpx.NewPageMeta(page, total))
}
