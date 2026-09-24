// HTTP handlers for /api/v1/notifications/*.
package notifications

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/skycode/ojt-management/backend/internal/auth"
	"github.com/skycode/ojt-management/backend/internal/platform/db"
	"github.com/skycode/ojt-management/backend/internal/platform/httpx"
)

type Handler struct {
	pool *db.Pool
}

func NewHandler(pool *db.Pool) *Handler { return &Handler{pool: pool} }

// List: GET /notifications?unread=true&page=
func (h *Handler) List(c fiber.Ctx) error {
	actor := auth.ActorOf(c)
	page := httpx.ParsePage(c)
	items, total, unread, err := List(c.Context(), h.pool, actor.ID,
		c.Query("unread") == "true", page)
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	meta := httpx.NewPageMeta(page, total)
	return httpx.OKMeta(c, fiber.Map{
		"items": items, "unread_count": unread,
	}, meta)
}

// Read: POST /notifications/{id}/read
func (h *Handler) Read(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, httpx.ErrNotFound)
	}
	if err := MarkRead(c.Context(), h.pool, auth.ActorOf(c).ID, id); err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, fiber.Map{"read": true})
}

// ReadAll: POST /notifications/read-all
func (h *Handler) ReadAll(c fiber.Ctx) error {
	if err := MarkAllRead(c.Context(), h.pool, auth.ActorOf(c).ID); err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	return httpx.OK(c, fiber.Map{"read": true})
}
