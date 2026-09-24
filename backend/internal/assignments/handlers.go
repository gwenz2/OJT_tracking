// HTTP handlers for /api/v1/staff/assignments/*.
package assignments

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/skycode/ojt-management/backend/internal/audit"
	"github.com/skycode/ojt-management/backend/internal/auth"
	"github.com/skycode/ojt-management/backend/internal/platform/db"
	"github.com/skycode/ojt-management/backend/internal/platform/httpx"
)

type Handler struct {
	pool *db.Pool
}

func NewHandler(pool *db.Pool) *Handler { return &Handler{pool: pool} }

// scopeFor resolves the caller's trainee-scope filter. Coordinators are
// restricted; admins get nil (unfiltered).
func (h *Handler) scopeFor(c fiber.Ctx) ([]uuid.UUID, error) {
	return auth.ScopedTraineeIDs(c.Context(), h.pool, auth.ActorOf(c))
}

// checkScope verifies a single trainee is within the caller's scope.
func (h *Handler) checkScope(c fiber.Ctx, traineeID uuid.UUID) error {
	ok, err := auth.CanAccessTrainee(c.Context(), h.pool, auth.ActorOf(c), traineeID)
	if err != nil {
		return httpx.Internal(err)
	}
	if !ok {
		return httpx.ErrOutOfScope
	}
	return nil
}

// List: GET /staff/assignments?trainee_id=&site_id=&status=&page=&page_size=
func (h *Handler) List(c fiber.Ctx) error {
	page := httpx.ParsePage(c)
	scope, err := h.scopeFor(c)
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	items, total, err := List(c.Context(), h.pool, page, scope,
		c.Query("trainee_id"), c.Query("site_id"), c.Query("status"))
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	return httpx.OKMeta(c, items, httpx.NewPageMeta(page, total))
}

// Create: POST /staff/assignments
func (h *Handler) Create(c fiber.Ctx) error {
	var in CreateInput
	if err := c.Bind().JSON(&in); err != nil {
		return httpx.Fail(c, httpx.ValidationMsg("Malformed request body."))
	}
	if tid, err := uuid.Parse(in.TraineeID); err == nil {
		if err := h.checkScope(c, tid); err != nil {
			return httpx.Fail(c, err)
		}
	}
	a, err := Create(c.Context(), h.pool, in, auth.ActorOf(c).ID)
	if err != nil {
		return httpx.Fail(c, err)
	}
	actor := auth.ActorOf(c)
	audit.Record(c.Context(), h.pool, &actor.ID, "assignment.created", "ojt_assignment", &a.ID,
		httpx.RequestID(c), map[string]any{"trainee_id": a.TraineeID, "site_id": a.SiteID})
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": a})
}

// GetOne: GET /staff/assignments/{id}
func (h *Handler) GetOne(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, httpx.ErrNotFound)
	}
	a, err := Get(c.Context(), h.pool, id)
	if err != nil {
		return httpx.Fail(c, err)
	}
	if err := h.checkScope(c, a.TraineeID); err != nil {
		// Hide existence for out-of-scope coordinators.
		if aerr, ok := err.(*httpx.APIError); ok && aerr.Code == "OUT_OF_SCOPE" {
			return httpx.Fail(c, httpx.ErrNotFound)
		}
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, a)
}

// Update: PATCH /staff/assignments/{id}
func (h *Handler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, httpx.ErrNotFound)
	}
	existing, err := Get(c.Context(), h.pool, id)
	if err != nil {
		return httpx.Fail(c, err)
	}
	if err := h.checkScope(c, existing.TraineeID); err != nil {
		if aerr, ok := err.(*httpx.APIError); ok && aerr.Code == "OUT_OF_SCOPE" {
			return httpx.Fail(c, httpx.ErrNotFound)
		}
		return httpx.Fail(c, err)
	}
	var in PatchInput
	if err := c.Bind().JSON(&in); err != nil {
		return httpx.Fail(c, httpx.ValidationMsg("Malformed request body."))
	}
	a, err := Update(c.Context(), h.pool, id, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	actor := auth.ActorOf(c)
	audit.Record(c.Context(), h.pool, &actor.ID, "assignment.updated", "ojt_assignment", &a.ID,
		httpx.RequestID(c), map[string]any{"trainee_id": a.TraineeID})
	return httpx.OK(c, a)
}
