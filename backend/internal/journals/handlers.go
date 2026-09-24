// HTTP handlers for /api/v1/journals/* and /api/v1/staff/journals/*.
package journals

import (
	"errors"
	"io"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/skycode/ojt-management/backend/internal/auth"
	"github.com/skycode/ojt-management/backend/internal/platform/db"
	"github.com/skycode/ojt-management/backend/internal/platform/httpx"
	"github.com/skycode/ojt-management/backend/internal/platform/storage"
)

type Handler struct {
	pool     *db.Pool
	store    *storage.Store
	maxBytes int64
}

func NewHandler(pool *db.Pool, store *storage.Store, maxUploadBytes int64) *Handler {
	return &Handler{pool: pool, store: store, maxBytes: maxUploadBytes}
}

// authorize returns the journal if the actor may see it: owner trainee,
// in-scope coordinator, or admin. Out-of-scope → 404 (no existence leak).
func (h *Handler) authorize(c fiber.Ctx, journalID uuid.UUID) (*Journal, error) {
	j, traineeID, err := Load(c.Context(), h.pool, journalID)
	if err != nil {
		return nil, err
	}
	actor := auth.ActorOf(c)
	if actor.Role == "trainee" {
		// Owner check: trainee_id → user_id compare.
		var ownerUser uuid.UUID
		if err := h.pool.QueryRow(c.Context(),
			`SELECT user_id FROM trainee_profiles WHERE id = $1`, traineeID).Scan(&ownerUser); err != nil {
			return nil, err
		}
		if ownerUser != actor.ID {
			return nil, httpx.ErrNotFound
		}
		return j, nil
	}
	ok, err := auth.CanAccessTrainee(c.Context(), h.pool, actor, traineeID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, httpx.ErrNotFound
	}
	return j, nil
}

// authorizeOwner is authorize + trainee-owner check for mutations.
func (h *Handler) authorizeOwner(c fiber.Ctx, journalID uuid.UUID) (*Journal, error) {
	actor := auth.ActorOf(c)
	if actor.Role != "trainee" {
		return nil, httpx.ErrForbidden
	}
	return h.authorize(c, journalID)
}

// Get: GET /journals/{id}
func (h *Handler) Get(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, httpx.ErrNotFound)
	}
	j, err := h.authorize(c, id)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, j)
}

// SaveDraft: PUT /journals/{id}/draft {narrative}
func (h *Handler) SaveDraft(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, httpx.ErrNotFound)
	}
	if _, err := h.authorizeOwner(c, id); err != nil {
		return httpx.Fail(c, err)
	}
	var in struct {
		Narrative string `json:"narrative"`
	}
	if err := c.Bind().JSON(&in); err != nil {
		return httpx.Fail(c, httpx.ValidationMsg("Malformed request body."))
	}
	if err := SaveDraft(c.Context(), h.pool, id, in.Narrative); err != nil {
		return httpx.Fail(c, err)
	}
	j, _, err := Load(c.Context(), h.pool, id)
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	return httpx.OK(c, j)
}

// Submit: POST /journals/{id}/submit {narrative}
func (h *Handler) Submit(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, httpx.ErrNotFound)
	}
	if _, err := h.authorizeOwner(c, id); err != nil {
		return httpx.Fail(c, err)
	}
	var in struct {
		Narrative string `json:"narrative"`
	}
	if err := c.Bind().JSON(&in); err != nil {
		return httpx.Fail(c, httpx.ValidationMsg("Malformed request body."))
	}
	if err := Submit(c.Context(), h.pool, id, auth.ActorOf(c).ID, in.Narrative); err != nil {
		return httpx.Fail(c, err)
	}
	j, _, err := Load(c.Context(), h.pool, id)
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	return httpx.OK(c, j)
}

// AddEvidence: POST /journals/{id}/evidence (multipart "photo")
func (h *Handler) AddEvidence(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, httpx.ErrNotFound)
	}
	if _, err := h.authorizeOwner(c, id); err != nil {
		return httpx.Fail(c, err)
	}
	fh, err := c.FormFile("photo")
	if err != nil {
		return httpx.Fail(c, httpx.Validation(map[string]string{"photo": "An image is required."}))
	}
	f, err := fh.Open()
	if err != nil {
		return httpx.Fail(c, httpx.New(415, "INVALID_IMAGE", "Could not read the file."))
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, 64<<20))
	if err != nil {
		return httpx.Fail(c, httpx.New(415, "INVALID_IMAGE", "Could not read the file."))
	}
	item, err := AddEvidence(c.Context(), h.pool, h.store, id, raw, h.maxBytes)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": item})
}

// DeleteEvidence: DELETE /journals/{id}/evidence/{eid}
func (h *Handler) DeleteEvidence(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, httpx.ErrNotFound)
	}
	eid, err := uuid.Parse(c.Params("eid"))
	if err != nil {
		return httpx.Fail(c, httpx.ErrNotFound)
	}
	if _, err := h.authorizeOwner(c, id); err != nil {
		return httpx.Fail(c, err)
	}
	if err := DeleteEvidence(c.Context(), h.pool, h.store, id, eid); err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, fiber.Map{"deleted": true})
}

// EvidenceImage: GET /journals/{id}/evidence/{eid}/image → presigned redirect.
func (h *Handler) EvidenceImage(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, httpx.ErrNotFound)
	}
	eid, err := uuid.Parse(c.Params("eid"))
	if err != nil {
		return httpx.Fail(c, httpx.ErrNotFound)
	}
	if _, err := h.authorize(c, id); err != nil {
		return httpx.Fail(c, err)
	}
	key, err := EvidenceImageKey(c.Context(), h.pool, id, eid)
	if errors.Is(err, pgx.ErrNoRows) {
		return httpx.Fail(c, httpx.ErrNotFound)
	}
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	url, err := h.store.PresignGet(c.Context(), key)
	if err != nil {
		return httpx.Fail(c, httpx.New(503, "STORAGE_UNAVAILABLE", "Storage unavailable.").WithInternal(err))
	}
	// Presigned URLs expire — never let the browser cache the redirect.
	c.Set("Cache-Control", "no-store")
	return c.Redirect().To(url)
}

// ---- staff -------------------------------------------------------------------

// Queue: GET /staff/journals?status=&page=
func (h *Handler) Queue(c fiber.Ctx) error {
	scope, err := auth.ScopedTraineeIDs(c.Context(), h.pool, auth.ActorOf(c))
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	page := httpx.ParsePage(c)
	items, total, err := ListQueue(c.Context(), h.pool, scope, c.Query("status"), page)
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	return httpx.OKMeta(c, items, httpx.NewPageMeta(page, total))
}

// Review: POST /staff/journals/{id}/review {decision, comment?}
func (h *Handler) Review(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, httpx.ErrNotFound)
	}
	// Scope check via Load → authorize (404s when out of scope).
	if _, err := h.authorize(c, id); err != nil {
		return httpx.Fail(c, err)
	}
	var in struct {
		Decision string  `json:"decision"`
		Comment  *string `json:"comment"`
	}
	if err := c.Bind().JSON(&in); err != nil {
		return httpx.Fail(c, httpx.ValidationMsg("Malformed request body."))
	}
	if err := Review(c.Context(), h.pool, id, auth.ActorOf(c).ID,
		strings.ToLower(in.Decision), in.Comment); err != nil {
		return httpx.Fail(c, err)
	}
	j, _, err := Load(c.Context(), h.pool, id)
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	return httpx.OK(c, j)
}
