// Daily journal workflow. One journal per completed attendance session;
// DRAFT → SUBMITTED → REVIEWED, with SUBMITTED → NEEDS_REVISION → SUBMITTED
// revision loop. Review history is append-only (spec §3.9–3.11).
package journals

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/skycode/ojt-management/backend/internal/audit"
	"github.com/skycode/ojt-management/backend/internal/notifications"
	"github.com/skycode/ojt-management/backend/internal/platform/db"
	"github.com/skycode/ojt-management/backend/internal/platform/httpx"
)

// narrativeMinLen is the "meaningful narrative" floor per contract §5.
const narrativeMinLen = 20

// maxEvidencePerJournal caps supporting images per journal.
const maxEvidencePerJournal = 5

// EnsureDraftTx creates the draft journal for a completed attendance inside
// the caller's transaction. Idempotent via the UNIQUE(attendance_id) index.
func EnsureDraftTx(ctx context.Context, tx pgx.Tx, attendanceID, traineeID uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID
	err := tx.QueryRow(ctx, `
		INSERT INTO daily_journals (attendance_id, trainee_id, status)
		VALUES ($1,$2,'draft')
		ON CONFLICT (attendance_id) DO NOTHING
		RETURNING id`, attendanceID, traineeID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		err = tx.QueryRow(ctx,
			`SELECT id FROM daily_journals WHERE attendance_id = $1`, attendanceID).Scan(&id)
	}
	return id, err
}

// ---- views ------------------------------------------------------------------

type EvidenceItem struct {
	ID        uuid.UUID `json:"id"`
	MimeType  string    `json:"mime_type"`
	SizeBytes int64     `json:"size_bytes"`
	Width     int       `json:"width_px"`
	Height    int       `json:"height_px"`
	CreatedAt time.Time `json:"created_at"`
}

type ReviewItem struct {
	ID        uuid.UUID `json:"id"`
	Decision  string    `json:"decision"`
	Comment   *string   `json:"comment"`
	Reviewer  string    `json:"reviewer_name"`
	CreatedAt time.Time `json:"created_at"`
}

type AttendanceSummary struct {
	Date            string     `json:"date"`
	SiteName        string     `json:"site_name"`
	TimeInAt        time.Time  `json:"time_in_at"`
	TimeOutAt       *time.Time `json:"time_out_at"`
	CreditedMinutes *int       `json:"credited_minutes"`
}

type Journal struct {
	ID            uuid.UUID         `json:"id"`
	Status        string            `json:"status"`
	Attendance    AttendanceSummary `json:"attendance"`
	AttendanceID  uuid.UUID         `json:"attendance_id"`
	Narrative     *string           `json:"narrative"`
	Evidence      []EvidenceItem    `json:"evidence"`
	LatestReview  *ReviewItem       `json:"latest_review"`
	ReviewHistory []ReviewItem      `json:"review_history"`
	SubmittedAt   *time.Time        `json:"submitted_at"`
	ReviewedAt    *time.Time        `json:"reviewed_at"`
	RevisionCount int               `json:"revision_count"`
	TraineeName   string            `json:"trainee_name,omitempty"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

type QueueItem struct {
	ID            uuid.UUID  `json:"id"`
	Status        string     `json:"status"`
	TraineeName   string     `json:"trainee_name"`
	TraineeID     uuid.UUID  `json:"trainee_id"`
	StudentNumber string     `json:"student_number"`
	SiteName      string     `json:"site_name"`
	Date          string     `json:"date"`
	SubmittedAt   *time.Time `json:"submitted_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// ---- queries ----------------------------------------------------------------

// Load fetches a journal with attendance summary, evidence, and reviews.
func Load(ctx context.Context, pool *db.Pool, journalID uuid.UUID) (*Journal, uuid.UUID, error) {
	var j Journal
	var traineeID uuid.UUID
	err := pool.QueryRow(ctx, `
		SELECT j.id, j.status, j.narrative, j.submitted_at, j.reviewed_at,
		       j.revision_count, j.updated_at, j.attendance_id, j.trainee_id,
		       u.display_name,
		       s.attendance_date::text, si.name, s.effective_time_in_at,
		       s.effective_time_out_at, s.credited_minutes
		FROM daily_journals j
		JOIN trainee_profiles tp ON tp.id = j.trainee_id
		JOIN users u ON u.id = tp.user_id
		JOIN attendance_sessions s ON s.id = j.attendance_id
		JOIN ojt_assignments a ON a.id = s.assignment_id
		JOIN ojt_sites si ON si.id = a.site_id
		WHERE j.id = $1`, journalID).
		Scan(&j.ID, &j.Status, &j.Narrative, &j.SubmittedAt, &j.ReviewedAt,
			&j.RevisionCount, &j.UpdatedAt, &j.AttendanceID, &traineeID,
			&j.TraineeName,
			&j.Attendance.Date, &j.Attendance.SiteName, &j.Attendance.TimeInAt,
			&j.Attendance.TimeOutAt, &j.Attendance.CreditedMinutes)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, uuid.Nil, httpx.ErrNotFound
	}
	if err != nil {
		return nil, uuid.Nil, err
	}

	evRows, err := pool.Query(ctx, `
		SELECT id, mime_type, size_bytes, width_px, height_px, created_at
		FROM journal_evidence WHERE journal_id = $1 ORDER BY created_at`, journalID)
	if err != nil {
		return nil, uuid.Nil, err
	}
	defer evRows.Close()
	j.Evidence = []EvidenceItem{}
	for evRows.Next() {
		var e EvidenceItem
		if err := evRows.Scan(&e.ID, &e.MimeType, &e.SizeBytes, &e.Width, &e.Height, &e.CreatedAt); err != nil {
			return nil, uuid.Nil, err
		}
		j.Evidence = append(j.Evidence, e)
	}

	rvRows, err := pool.Query(ctx, `
		SELECT r.id, r.decision, r.comment, u.display_name, r.created_at
		FROM journal_reviews r JOIN users u ON u.id = r.reviewer_user_id
		WHERE r.journal_id = $1 ORDER BY r.created_at DESC`, journalID)
	if err != nil {
		return nil, uuid.Nil, err
	}
	defer rvRows.Close()
	j.ReviewHistory = []ReviewItem{}
	for rvRows.Next() {
		var r ReviewItem
		if err := rvRows.Scan(&r.ID, &r.Decision, &r.Comment, &r.Reviewer, &r.CreatedAt); err != nil {
			return nil, uuid.Nil, err
		}
		j.ReviewHistory = append(j.ReviewHistory, r)
	}
	if len(j.ReviewHistory) > 0 {
		j.LatestReview = &j.ReviewHistory[0]
	}
	return &j, traineeID, nil
}

// ListQueue is the staff queue: journals needing attention, scope-filtered.
func ListQueue(ctx context.Context, pool *db.Pool, scope []uuid.UUID, status string, p httpx.Page) ([]QueueItem, int64, error) {
	where := `WHERE ($1::uuid[] IS NULL OR j.trainee_id = ANY($1))
		AND ($2::text IS NULL OR j.status = $2)`
	var scopeArg any
	if scope != nil {
		scopeArg = scope
	}
	var statusArg any
	if status != "" {
		statusArg = status
	}

	var total int64
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM daily_journals j `+where, scopeArg, statusArg).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := pool.Query(ctx, `
		SELECT j.id, j.status, u.display_name, j.trainee_id, tp.student_number,
		       si.name, s.attendance_date::text, j.submitted_at, j.updated_at
		FROM daily_journals j
		JOIN trainee_profiles tp ON tp.id = j.trainee_id
		JOIN users u ON u.id = tp.user_id
		JOIN attendance_sessions s ON s.id = j.attendance_id
		JOIN ojt_assignments a ON a.id = s.assignment_id
		JOIN ojt_sites si ON si.id = a.site_id
		`+where+`
		ORDER BY CASE j.status WHEN 'submitted' THEN 0 WHEN 'needs_revision' THEN 1 ELSE 2 END,
		         j.updated_at DESC
		LIMIT $3 OFFSET $4`, scopeArg, statusArg, p.Limit(), p.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := []QueueItem{}
	for rows.Next() {
		var it QueueItem
		if err := rows.Scan(&it.ID, &it.Status, &it.TraineeName, &it.TraineeID,
			&it.StudentNumber, &it.SiteName, &it.Date, &it.SubmittedAt, &it.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, it)
	}
	return items, total, rows.Err()
}

// ---- mutations ----------------------------------------------------------------

// editableErr when the journal is not in draft/needs_revision.
var errNotEditable = httpx.Conflict("JOURNAL_NOT_EDITABLE", "This journal is no longer editable.")

// SaveDraft persists narrative while the journal is editable.
func SaveDraft(ctx context.Context, pool *db.Pool, journalID uuid.UUID, narrative string) error {
	res, err := pool.Exec(ctx, `
		UPDATE daily_journals SET narrative = $2, updated_at = now()
		WHERE id = $1 AND status IN ('draft','needs_revision')`,
		journalID, narrative)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errNotEditable
	}
	return nil
}

// Submit transitions draft|needs_revision → submitted after validation.
// Locks the row so a concurrent review cannot interleave mid-transition.
func Submit(ctx context.Context, pool *db.Pool, journalID, actorID uuid.UUID, narrative string) error {
	narrative = strings.TrimSpace(narrative)
	if len(narrative) < narrativeMinLen {
		return httpx.Unprocessable("NARRATIVE_REQUIRED",
			"Write at least a couple of sentences about your day.")
	}
	return pool.InTx(ctx, func(tx pgx.Tx) error {
		var status string
		err := tx.QueryRow(ctx,
			`SELECT status FROM daily_journals WHERE id = $1 FOR UPDATE`, journalID).Scan(&status)
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		}
		if err != nil {
			return err
		}
		if status != "draft" && status != "needs_revision" {
			return errNotEditable
		}
		_, err = tx.Exec(ctx, `
			UPDATE daily_journals SET narrative = $2, status = 'submitted',
			    submitted_at = now(), updated_at = now()
			WHERE id = $1`, journalID, narrative)
		if err != nil {
			return err
		}
		audit.Record(ctx, tx, &actorID, "journal.submitted", "daily_journal", &journalID, "", nil)
		return nil
	})
}

// Review applies a staff decision under a row lock.
// decision ∈ {reviewed, needs_revision}; needs_revision requires a comment.
func Review(ctx context.Context, pool *db.Pool, journalID, reviewerID uuid.UUID, decision string, comment *string) error {
	if decision != "reviewed" && decision != "needs_revision" {
		return httpx.Validation(map[string]string{"decision": "Must be reviewed or needs_revision."})
	}
	if decision == "needs_revision" && (comment == nil || strings.TrimSpace(*comment) == "") {
		return httpx.Unprocessable("REVIEW_COMMENT_REQUIRED",
			"A comment is required when requesting revision.")
	}
	return pool.InTx(ctx, func(tx pgx.Tx) error {
		var status string
		err := tx.QueryRow(ctx,
			`SELECT status FROM daily_journals WHERE id = $1 FOR UPDATE`, journalID).Scan(&status)
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		}
		if err != nil {
			return err
		}
		if status != "submitted" {
			return httpx.Conflict("JOURNAL_NOT_SUBMITTED",
				"Only submitted journals can be reviewed.")
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO journal_reviews (journal_id, reviewer_user_id, decision, comment)
			VALUES ($1,$2,$3,$4)`, journalID, reviewerID, decision, comment); err != nil {
			return err
		}
		incRevision := decision == "needs_revision"
		_, err = tx.Exec(ctx, `
			UPDATE daily_journals SET status = $2, reviewed_at = now(), updated_at = now(),
			    revision_count = revision_count + CASE WHEN $3 THEN 1 ELSE 0 END
			WHERE id = $1`, journalID, decision, incRevision)
		if err != nil {
			return err
		}

		// Notify the trainee of the decision (deduped per decision event).
		var traineeUser uuid.UUID
		if err := tx.QueryRow(ctx, `
			SELECT tp.user_id FROM daily_journals j
			JOIN trainee_profiles tp ON tp.id = j.trainee_id
			WHERE j.id = $1`, journalID).Scan(&traineeUser); err != nil {
			return err
		}
		title, body := "Journal reviewed", "Your daily journal was reviewed."
		if decision == "needs_revision" {
			title = "Journal needs revision"
			body = "Your coordinator requested changes to your daily journal."
		}
		var revCount int
		_ = tx.QueryRow(ctx,
			`SELECT count(*) FROM journal_reviews WHERE journal_id = $1`, journalID).Scan(&revCount)
		if _, err := notifications.CreateTx(ctx, tx, traineeUser,
			fmt.Sprintf("journal:%s:review:%d", journalID, revCount),
			"journal_reviewed", title, body, "daily_journal", &journalID); err != nil {
			return err
		}

		audit.Record(ctx, tx, &reviewerID, "journal."+decision, "daily_journal", &journalID, "", nil)
		return nil
	})
}

// journalRow is the locked state needed by evidence mutations.
func lockEditable(ctx context.Context, tx pgx.Tx, journalID uuid.UUID) error {
	var status string
	err := tx.QueryRow(ctx,
		`SELECT status FROM daily_journals WHERE id = $1 FOR UPDATE`, journalID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return httpx.ErrNotFound
	}
	if err != nil {
		return err
	}
	if status != "draft" && status != "needs_revision" {
		return errNotEditable
	}
	return nil
}
