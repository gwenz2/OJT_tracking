// HTTP handlers for /api/v1/staff/trainees/*.
package trainees

import (
	"io"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/skycode/ojt-management/backend/internal/audit"
	"github.com/skycode/ojt-management/backend/internal/auth"
	"github.com/skycode/ojt-management/backend/internal/platform/db"
	"github.com/skycode/ojt-management/backend/internal/platform/httpx"
)

type Handler struct {
	pool  *db.Pool
	store *ImportStore
}

func NewHandler(pool *db.Pool) *Handler {
	return &Handler{pool: pool, store: NewImportStore()}
}

func (h *Handler) scopeFor(c fiber.Ctx) ([]uuid.UUID, error) {
	return auth.ScopedTraineeIDs(c.Context(), h.pool, auth.ActorOf(c))
}

// List: GET /staff/trainees?q=&status=&site_id=&page=&page_size=
func (h *Handler) List(c fiber.Ctx) error {
	page := httpx.ParsePage(c)
	scope, err := h.scopeFor(c)
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	items, total, err := List(c.Context(), h.pool, page, scope,
		c.Query("q"), c.Query("status"), c.Query("site_id"))
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	return httpx.OKMeta(c, items, httpx.NewPageMeta(page, total))
}

// Create: POST /staff/trainees — creates user+profile; coordinator creators
// automatically scope the new trainee to themselves.
func (h *Handler) Create(c fiber.Ctx) error {
	var in CreateInput
	if err := c.Bind().JSON(&in); err != nil {
		return httpx.Fail(c, httpx.ValidationMsg("Malformed request body."))
	}
	t, tempPassword, err := Create(c.Context(), h.pool, in, auth.ActorOf(c))
	if err != nil {
		return httpx.Fail(c, err)
	}
	actor := auth.ActorOf(c)
	audit.Record(c.Context(), h.pool, &actor.ID, "trainee.created", "trainee_profile", &t.ID,
		httpx.RequestID(c), nil)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": fiber.Map{
		"trainee":            t,
		"temporary_password": tempPassword,
	}})
}

// GetOne: GET /staff/trainees/{id} — scope-enforced; out-of-scope returns 404.
func (h *Handler) GetOne(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, httpx.ErrNotFound)
	}
	if err := h.checkScope(c, id); err != nil {
		return httpx.Fail(c, err)
	}
	t, err := Get(c.Context(), h.pool, id)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, t)
}

// Update: PATCH /staff/trainees/{id}
func (h *Handler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, httpx.ErrNotFound)
	}
	if err := h.checkScope(c, id); err != nil {
		return httpx.Fail(c, err)
	}
	var in PatchInput
	if err := c.Bind().JSON(&in); err != nil {
		return httpx.Fail(c, httpx.ValidationMsg("Malformed request body."))
	}
	t, err := Update(c.Context(), h.pool, id, in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	actor := auth.ActorOf(c)
	audit.Record(c.Context(), h.pool, &actor.ID, "trainee.updated", "trainee_profile", &t.ID,
		httpx.RequestID(c), map[string]any{"account_status": t.AccountStatus})
	return httpx.OK(c, t)
}

// ImportValidate: POST /staff/trainees/import/validate (multipart file "file").
func (h *Handler) ImportValidate(c fiber.Ctx) error {
	fh, err := c.FormFile("file")
	if err != nil {
		return httpx.Fail(c, httpx.Validation(map[string]string{"file": "A CSV file is required."}))
	}
	if !strings.HasSuffix(strings.ToLower(fh.Filename), ".csv") {
		return httpx.Fail(c, httpx.Validation(map[string]string{"file": "File must be a .csv."}))
	}
	f, err := fh.Open()
	if err != nil {
		return httpx.Fail(c, httpx.ValidationMsg("Could not read the uploaded file."))
	}
	defer f.Close()

	limited := io.LimitReader(f, 1<<20) // 1 MiB CSV cap
	rows, results, err := ParseAndValidate(c.Context(), h.pool, limited)
	if err != nil {
		return httpx.Fail(c, err)
	}
	token, err := Stage(h.store, rows, auth.ActorOf(c))
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	preview := ImportPreview{Rows: results, ImportToken: token}
	for _, r := range results {
		if r.Status == "valid" {
			preview.ValidRows++
		} else {
			preview.InvalidRows++
		}
	}
	return httpx.OK(c, preview)
}

// ImportCommit: POST /staff/trainees/import/commit {import_token}
func (h *Handler) ImportCommit(c fiber.Ctx) error {
	var body struct {
		ImportToken string `json:"import_token"`
	}
	if err := c.Bind().JSON(&body); err != nil {
		return httpx.Fail(c, httpx.ValidationMsg("Malformed request body."))
	}
	res, err := Commit(c.Context(), h.pool, h.store, body.ImportToken)
	if err != nil {
		return httpx.Fail(c, err)
	}
	actor := auth.ActorOf(c)
	audit.Record(c.Context(), h.pool, &actor.ID, "trainee.imported", "trainee_profile", nil,
		httpx.RequestID(c), map[string]any{"created": res.Created})
	return httpx.OK(c, res)
}

// checkScope hides existence (404) for out-of-scope coordinators.
func (h *Handler) checkScope(c fiber.Ctx, traineeID uuid.UUID) error {
	ok, err := auth.CanAccessTrainee(c.Context(), h.pool, auth.ActorOf(c), traineeID)
	if err != nil {
		return httpx.Internal(err)
	}
	if !ok {
		return httpx.ErrNotFound
	}
	return nil
}
