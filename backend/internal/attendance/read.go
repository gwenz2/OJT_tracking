// Attendance read endpoints: today, history, detail, evidence media.
package attendance

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/skycode/ojt-management/backend/internal/auth"
	"github.com/skycode/ojt-management/backend/internal/platform/db"
	"github.com/skycode/ojt-management/backend/internal/platform/httpx"
	"github.com/skycode/ojt-management/backend/internal/platform/storage"
)

type ReadHandler struct {
	pool  *db.Pool
	store *storage.Store
}

func NewReadHandler(pool *db.Pool, store *storage.Store) *ReadHandler {
	return &ReadHandler{pool: pool, store: store}
}

// ---- GET /attendance/today -------------------------------------------------

func (h *ReadHandler) Today(c fiber.Ctx) error {
	actor := auth.ActorOf(c)
	tz, err := h.timezone(c.Context())
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	today := MustLocalDate(time.Now(), tz)

	var v AttendanceView
	var flags []string
	err = h.pool.QueryRow(c.Context(), `
		SELECT s.id, s.attendance_date::text, s.status,
		       s.effective_time_in_at, s.effective_time_out_at, s.credited_minutes,
		       s.flag_codes
		FROM attendance_sessions s
		JOIN trainee_profiles tp ON tp.id = s.trainee_id
		WHERE tp.user_id = $1 AND s.attendance_date = $2::date`, actor.ID, today).
		Scan(&v.ID, &v.Date, &v.Status, &v.OfficialTimeInAt, &v.OfficialTimeOutAt,
			&v.CreditedMinutes, &flags)
	if errors.Is(err, pgx.ErrNoRows) {
		return httpx.OK(c, nil)
	}
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	v.Flags = flags
	return httpx.OK(c, v)
}

// ---- GET /attendance/history ------------------------------------------------

type historyItem struct {
	ID              uuid.UUID  `json:"id"`
	Date            string     `json:"date"`
	Status          string     `json:"status"`
	SiteName        string     `json:"site_name"`
	TimeInAt        time.Time  `json:"time_in_at"`
	TimeOutAt       *time.Time `json:"time_out_at"`
	CreditedMinutes *int       `json:"credited_minutes"`
	Flags           []string   `json:"flags"`
	JournalStatus   *string    `json:"journal_status"`
}

func (h *ReadHandler) History(c fiber.Ctx) error {
	actor := auth.ActorOf(c)
	page := httpx.ParsePage(c)

	where := `WHERE tp.user_id = $1`
	args := []any{actor.ID}
	n := 2
	if v := c.Query("from"); v != "" {
		where += fmt.Sprintf(` AND s.attendance_date >= $%d::date`, n)
		args = append(args, v)
		n++
	}
	if v := c.Query("to"); v != "" {
		where += fmt.Sprintf(` AND s.attendance_date <= $%d::date`, n)
		args = append(args, v)
		n++
	}
	if v := c.Query("status"); v != "" {
		where += fmt.Sprintf(` AND s.status = $%d`, n)
		args = append(args, v)
		n++
	}

	var total int64
	countQ := `SELECT count(*) FROM attendance_sessions s
		JOIN trainee_profiles tp ON tp.id = s.trainee_id ` + where
	if err := h.pool.QueryRow(c.Context(), countQ, args...).Scan(&total); err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}

	q := `SELECT s.id, s.attendance_date::text, s.status, si.name,
	             s.effective_time_in_at, s.effective_time_out_at, s.credited_minutes,
	             s.flag_codes, j.status
	      FROM attendance_sessions s
	      JOIN trainee_profiles tp ON tp.id = s.trainee_id
	      JOIN ojt_assignments a ON a.id = s.assignment_id
	      JOIN ojt_sites si ON si.id = a.site_id
	      LEFT JOIN daily_journals j ON j.attendance_id = s.id
	      ` + where + ` ORDER BY s.attendance_date DESC, s.created_at DESC
	      LIMIT ` + fmt.Sprintf("%d", page.Limit()) + ` OFFSET ` + fmt.Sprintf("%d", page.Offset())
	rows, err := h.pool.Query(c.Context(), q, args...)
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	defer rows.Close()

	items := []historyItem{}
	for rows.Next() {
		var it historyItem
		if err := rows.Scan(&it.ID, &it.Date, &it.Status, &it.SiteName,
			&it.TimeInAt, &it.TimeOutAt, &it.CreditedMinutes, &it.Flags, &it.JournalStatus); err != nil {
			return httpx.Fail(c, httpx.Internal(err))
		}
		items = append(items, it)
	}
	return httpx.OKMeta(c, items, httpx.NewPageMeta(page, total))
}

// ---- GET /attendance/{id} ----------------------------------------------------

type evidenceRef struct {
	Action            string     `json:"action"`
	LocationStatus    string     `json:"location_status"`
	DistanceFromSiteM *float64   `json:"distance_from_site_m"`
	GPSAccuracyM      *float64   `json:"gps_accuracy_m"`
	ServerCapturedAt  time.Time  `json:"server_captured_at"`
	DeviceCapturedAt  *time.Time `json:"device_captured_at"`
	ExceptionReason   *string    `json:"location_exception_reason"`
	Width             int        `json:"width_px"`
	Height            int        `json:"height_px"`
}

type correctionRef struct {
	ID                uuid.UUID  `json:"id"`
	Type              string     `json:"type"`
	Status            string     `json:"status"`
	Reason            string     `json:"reason"`
	ProposedTimeInAt  *time.Time `json:"proposed_time_in_at"`
	ProposedTimeOutAt *time.Time `json:"proposed_time_out_at"`
	DecisionComment   *string    `json:"decision_comment"`
	RequestedAt       time.Time  `json:"requested_at"`
	DecidedAt         *time.Time `json:"decided_at"`
}

type detailResponse struct {
	ID                 uuid.UUID       `json:"id"`
	Date               string          `json:"date"`
	Status             string          `json:"status"`
	TraineeName        string          `json:"trainee_name"`
	SiteName           string          `json:"site_name"`
	OriginalTimeInAt   time.Time       `json:"original_time_in_at"`
	OriginalTimeOutAt  *time.Time      `json:"original_time_out_at"`
	EffectiveTimeInAt  time.Time       `json:"effective_time_in_at"`
	EffectiveTimeOutAt *time.Time      `json:"effective_time_out_at"`
	CreditedMinutes    *int            `json:"credited_minutes"`
	Flags              []string        `json:"flags"`
	Evidence           []evidenceRef   `json:"evidence"`
	Corrections        []correctionRef `json:"correction_history"`
	JournalID          *uuid.UUID      `json:"journal_id"`
}

// Detail authorizes: owner trainee, coordinator in scope of the trainee, admin.
func (h *ReadHandler) Detail(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, httpx.New(404, "ATTENDANCE_NOT_FOUND", "Attendance record not found."))
	}
	actor := auth.ActorOf(c)

	var (
		d         detailResponse
		traineeID uuid.UUID
		ownerUser uuid.UUID
	)
	err = h.pool.QueryRow(c.Context(), `
		SELECT s.id, s.attendance_date::text, s.status, u.display_name, si.name,
		       s.original_time_in_at, s.original_time_out_at,
		       s.effective_time_in_at, s.effective_time_out_at,
		       s.credited_minutes, s.flag_codes, s.trainee_id, tp.user_id,
		       (SELECT j.id FROM daily_journals j WHERE j.attendance_id = s.id)
		FROM attendance_sessions s
		JOIN trainee_profiles tp ON tp.id = s.trainee_id
		JOIN users u ON u.id = tp.user_id
		JOIN ojt_assignments a ON a.id = s.assignment_id
		JOIN ojt_sites si ON si.id = a.site_id
		WHERE s.id = $1`, id).
		Scan(&d.ID, &d.Date, &d.Status, &d.TraineeName, &d.SiteName,
			&d.OriginalTimeInAt, &d.OriginalTimeOutAt,
			&d.EffectiveTimeInAt, &d.EffectiveTimeOutAt,
			&d.CreditedMinutes, &d.Flags, &traineeID, &ownerUser, &d.JournalID)
	if errors.Is(err, pgx.ErrNoRows) {
		return httpx.Fail(c, httpx.New(404, "ATTENDANCE_NOT_FOUND", "Attendance record not found."))
	}
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}

	// Authorization: owner or in-scope staff. Out-of-scope → 404.
	if actor.ID != ownerUser {
		ok, err := auth.CanAccessTrainee(c.Context(), h.pool, actor, traineeID)
		if err != nil {
			return httpx.Fail(c, httpx.Internal(err))
		}
		if !ok {
			return httpx.Fail(c, httpx.New(404, "ATTENDANCE_NOT_FOUND", "Attendance record not found."))
		}
	}

	rows, err := h.pool.Query(c.Context(), `
		SELECT action, location_status, distance_from_site_m, gps_accuracy_m,
		       server_captured_at, device_captured_at, location_exception_reason,
		       width_px, height_px
		FROM attendance_evidence WHERE attendance_id = $1 ORDER BY server_captured_at`, id)
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	defer rows.Close()
	d.Evidence = []evidenceRef{}
	for rows.Next() {
		var e evidenceRef
		if err := rows.Scan(&e.Action, &e.LocationStatus, &e.DistanceFromSiteM, &e.GPSAccuracyM,
			&e.ServerCapturedAt, &e.DeviceCapturedAt, &e.ExceptionReason, &e.Width, &e.Height); err != nil {
			return httpx.Fail(c, httpx.Internal(err))
		}
		d.Evidence = append(d.Evidence, e)
	}

	// Correction history — the original/effective split is only meaningful
	// alongside the decisions that produced it (T116).
	cRows, err := h.pool.Query(c.Context(), `
		SELECT id, type, status, reason, proposed_time_in_at, proposed_time_out_at,
		       decision_comment, requested_at, decided_at
		FROM correction_requests WHERE attendance_id = $1 ORDER BY requested_at`, id)
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	defer cRows.Close()
	d.Corrections = []correctionRef{}
	for cRows.Next() {
		var cr correctionRef
		if err := cRows.Scan(&cr.ID, &cr.Type, &cr.Status, &cr.Reason,
			&cr.ProposedTimeInAt, &cr.ProposedTimeOutAt,
			&cr.DecisionComment, &cr.RequestedAt, &cr.DecidedAt); err != nil {
			return httpx.Fail(c, httpx.Internal(err))
		}
		d.Corrections = append(d.Corrections, cr)
	}
	return httpx.OK(c, d)
}

// ---- GET /attendance/{id}/evidence/{action}/image -----------------------------

// EvidenceImage redirects to a short-lived presigned URL for the evidence
// image. variant=watermarked (default) | original. Owner/in-scope staff only;
// the object key is never exposed to unauthorized callers.
func (h *ReadHandler) EvidenceImage(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, httpx.ErrNotFound)
	}
	action := c.Params("action")
	if action != "time_in" && action != "time_out" {
		return httpx.Fail(c, httpx.ErrNotFound)
	}
	variant := c.Query("variant", "watermarked")
	if variant != "watermarked" && variant != "original" {
		return httpx.Fail(c, httpx.ValidationMsg("variant must be watermarked or original."))
	}

	var traineeID, ownerUser uuid.UUID
	var key string
	col := "watermarked_object_key"
	if variant == "original" {
		col = "original_object_key"
	}
	err = h.pool.QueryRow(c.Context(), `
		SELECT s.trainee_id, tp.user_id, e.`+col+`
		FROM attendance_evidence e
		JOIN attendance_sessions s ON s.id = e.attendance_id
		JOIN trainee_profiles tp ON tp.id = s.trainee_id
		WHERE s.id = $1 AND e.action = $2`, id, action).
		Scan(&traineeID, &ownerUser, &key)
	if errors.Is(err, pgx.ErrNoRows) {
		return httpx.Fail(c, httpx.ErrNotFound)
	}
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}

	actor := auth.ActorOf(c)
	if actor.ID != ownerUser {
		ok, err := auth.CanAccessTrainee(c.Context(), h.pool, actor, traineeID)
		if err != nil {
			return httpx.Fail(c, httpx.Internal(err))
		}
		if !ok {
			return httpx.Fail(c, httpx.ErrNotFound)
		}
	}

	url, err := h.store.PresignGet(c.Context(), key)
	if err != nil {
		return httpx.Fail(c, storageUnavailable(err))
	}
	// Presigned URLs expire — never let the browser cache the redirect.
	c.Set("Cache-Control", "no-store")
	return c.Redirect().To(url)
}

func (h *ReadHandler) timezone(ctx context.Context) (string, error) {
	var tz string
	err := h.pool.QueryRow(ctx, `SELECT timezone FROM institution_settings LIMIT 1`).Scan(&tz)
	return tz, err
}
