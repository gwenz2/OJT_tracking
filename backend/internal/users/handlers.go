// HTTP handlers for /api/v1/admin/coordinators/*.
package users

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

func (h *Handler) List(c fiber.Ctx) error {
	page := httpx.ParsePage(c)
	items, total, err := ListCoordinators(c.Context(), h.pool, page, c.Query("q"))
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	return httpx.OKMeta(c, items, httpx.NewPageMeta(page, total))
}

func (h *Handler) Create(c fiber.Ctx) error {
	var in CreateCoordinatorInput
	if err := c.Bind().JSON(&in); err != nil {
		return httpx.Fail(c, httpx.ValidationMsg("Malformed request body."))
	}
	coord, tempPassword, err := CreateCoordinator(c.Context(), h.pool, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	actor := auth.ActorOf(c)
	audit.Record(c.Context(), h.pool, &actor.ID, "coordinator.created", "user", &coord.ID, httpx.RequestID(c), nil)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": fiber.Map{
		"coordinator":        coord,
		"temporary_password": tempPassword,
	}})
}

func (h *Handler) Patch(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, httpx.ErrNotFound)
	}
	var in PatchCoordinatorInput
	if err := c.Bind().JSON(&in); err != nil {
		return httpx.Fail(c, httpx.ValidationMsg("Malformed request body."))
	}
	coord, err := PatchCoordinator(c.Context(), h.pool, id, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	actor := auth.ActorOf(c)
	audit.Record(c.Context(), h.pool, &actor.ID, "coordinator.updated", "user", &coord.ID,
		httpx.RequestID(c), map[string]any{"account_status": coord.AccountStatus})
	return httpx.OK(c, coord)
}

func (h *Handler) SetScope(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, httpx.ErrNotFound)
	}
	var body struct {
		TraineeIDs []uuid.UUID `json:"trainee_ids"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return httpx.Fail(c, httpx.ValidationMsg("Malformed request body."))
	}
	if err := SetTraineeScope(c.Context(), h.pool, id, body.TraineeIDs); err != nil {
		return httpx.Fail(c, err)
	}
	actor := auth.ActorOf(c)
	audit.Record(c.Context(), h.pool, &actor.ID, "coordinator.scope_updated", "user", &id,
		httpx.RequestID(c), map[string]any{"trainee_count": len(body.TraineeIDs)})
	return httpx.OK(c, fiber.Map{"updated": true})
}
