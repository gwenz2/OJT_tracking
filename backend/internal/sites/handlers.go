// HTTP handlers for /api/v1/staff/sites/*.
package sites

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

// List: GET /staff/sites?q=&is_active=&page=&page_size=
func (h *Handler) List(c fiber.Ctx) error {
	page := httpx.ParsePage(c)
	var active *bool
	if v := c.Query("is_active"); v != "" {
		b := v == "true" || v == "1"
		active = &b
	}
	items, total, err := List(c.Context(), h.pool, page, c.Query("q"), active)
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	return httpx.OKMeta(c, items, httpx.NewPageMeta(page, total))
}

// Create: POST /staff/sites
func (h *Handler) Create(c fiber.Ctx) error {
	var in Input
	if err := c.Bind().JSON(&in); err != nil {
		return httpx.Fail(c, httpx.ValidationMsg("Malformed request body."))
	}
	s, err := Create(c.Context(), h.pool, in, auth.ActorOf(c).ID)
	if err != nil {
		return httpx.Fail(c, err)
	}
	audit.Record(c.Context(), h.pool, ptr(auth.ActorOf(c).ID), "site.created", "ojt_site", &s.ID, httpx.RequestID(c), nil)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": s})
}

// GetOne: GET /staff/sites/{id}
func (h *Handler) GetOne(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, httpx.ErrNotFound)
	}
	s, err := Get(c.Context(), h.pool, id)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, s)
}

// Update: PATCH /staff/sites/{id}
func (h *Handler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, httpx.ErrNotFound)
	}
	var in Input
	if err := c.Bind().JSON(&in); err != nil {
		return httpx.Fail(c, httpx.ValidationMsg("Malformed request body."))
	}
	s, err := Update(c.Context(), h.pool, id, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	audit.Record(c.Context(), h.pool, ptr(auth.ActorOf(c).ID), "site.updated", "ojt_site", &s.ID, httpx.RequestID(c), nil)
	return httpx.OK(c, s)
}

func ptr(v uuid.UUID) *uuid.UUID { return &v }
