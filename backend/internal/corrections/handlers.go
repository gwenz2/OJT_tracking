// HTTP handlers for correction requests + decisions.
package corrections

import (
	"time"

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

// traineeIDOf resolves the trainee profile id for the acting user.
func (h *Handler) traineeIDOf(c fiber.Ctx) (uuid.UUID, error) {
	var id uuid.UUID
	err := h.pool.QueryRow(c.Context(),
		`SELECT id FROM trainee_profiles WHERE user_id = $1`, auth.ActorOf(c).ID).Scan(&id)
	return id, err
}

// Create: POST /attendance/{attendance_id}/corrections
func (h *Handler) Create(c fiber.Ctx) error {
	attID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, httpx.ErrNotFound)
	}
	var in struct {
		Type              string     `json:"type"`
		Reason            string     `json:"reason"`
		ProposedTimeInAt  *time.Time `json:"proposed_time_in_at"`
		ProposedTimeOutAt *time.Time `json:"proposed_time_out_at"`
	}
	if err := c.Bind().JSON(&in); err != nil {
		return httpx.Fail(c, httpx.ValidationMsg("Malformed request body."))
	}
	traineeID, err := h.traineeIDOf(c)
	if err != nil {
		return httpx.Fail(c, httpx.ErrNotFound)
	}
	created, err := Request(c.Context(), h.pool, traineeID, auth.ActorOf(c).ID, RequestInput{
		AttendanceID:      attID,
		Type:              in.Type,
		Reason:            in.Reason,
		ProposedTimeInAt:  in.ProposedTimeInAt,
		ProposedTimeOutAt: in.ProposedTimeOutAt,
	})
	if err != nil {
		return httpx.Fail(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": created})
}

// Mine: GET /trainee/corrections
func (h *Handler) Mine(c fiber.Ctx) error {
	traineeID, err := h.traineeIDOf(c)
	if err != nil {
		return httpx.Fail(c, httpx.ErrNotFound)
	}
	page := httpx.ParsePage(c)
	items, total, err := ListMine(c.Context(), h.pool, traineeID, page)
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	return httpx.OKMeta(c, items, httpx.NewPageMeta(page, total))
}

// List: GET /staff/corrections?status=&type=&trainee_id=&from=&to=
func (h *Handler) List(c fiber.Ctx) error {
	scope, err := auth.ScopedTraineeIDs(c.Context(), h.pool, auth.ActorOf(c))
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	page := httpx.ParsePage(c)
	items, total, err := ListStaff(c.Context(), h.pool, scope,
		c.Query("status"), c.Query("type"), c.Query("trainee_id"),
		c.Query("from"), c.Query("to"), page)
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	return httpx.OKMeta(c, items, httpx.NewPageMeta(page, total))
}

// Decide: POST /staff/corrections/{id}/decision {decision, comment?}
func (h *Handler) Decide(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, httpx.ErrNotFound)
	}
	// Scope check before touching state — 404 when out of scope.
	traineeID, err := TraineeIDOf(c.Context(), h.pool, id)
	if err != nil {
		return httpx.Fail(c, err)
	}
	ok, err := auth.CanAccessTrainee(c.Context(), h.pool, auth.ActorOf(c), traineeID)
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	if !ok {
		return httpx.Fail(c, httpx.ErrNotFound)
	}
	var in struct {
		Decision string  `json:"decision"`
		Comment  *string `json:"comment"`
	}
	if err := c.Bind().JSON(&in); err != nil {
		return httpx.Fail(c, httpx.ValidationMsg("Malformed request body."))
	}
	if err := Decide(c.Context(), h.pool, id, auth.ActorOf(c).ID, in.Decision, in.Comment); err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, fiber.Map{"decision": in.Decision})
}
