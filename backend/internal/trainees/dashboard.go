// Trainee-facing dashboard: GET /api/v1/trainee/dashboard. Returns the
// active assignment progress, today's attendance/journal state, and the
// derived next_action the mobile shell should offer.
package trainees

import (
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/skycode/ojt-management/backend/internal/attendance"
	"github.com/skycode/ojt-management/backend/internal/auth"
	"github.com/skycode/ojt-management/backend/internal/platform/db"
	"github.com/skycode/ojt-management/backend/internal/platform/httpx"
)

type DashboardHandler struct {
	pool *db.Pool
}

func NewDashboardHandler(pool *db.Pool) *DashboardHandler {
	return &DashboardHandler{pool: pool}
}

type dashboardAssignment struct {
	ID               uuid.UUID `json:"id"`
	Site             siteRef   `json:"site"`
	RequiredMinutes  int       `json:"required_minutes"`
	CompletedMinutes int       `json:"completed_minutes"`
	RemainingMinutes int       `json:"remaining_minutes"`
	ProgressPercent  int       `json:"progress_percent"`
}

type siteRef struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type dashboardToday struct {
	AttendanceID     *uuid.UUID `json:"attendance_id"`
	AttendanceStatus *string    `json:"attendance_status"`
	TimeInAt         *time.Time `json:"time_in_at"`
	TimeOutAt        *time.Time `json:"time_out_at"`
	JournalID        *uuid.UUID `json:"journal_id"`
	JournalStatus    *string    `json:"journal_status"`
	NextAction       string     `json:"next_action"`
}

type dashboardResponse struct {
	Assignment          *dashboardAssignment `json:"assignment"`
	Today               dashboardToday       `json:"today"`
	UnreadNotifications int                  `json:"unread_notifications"`
}

// Get: GET /trainee/dashboard
func (h *DashboardHandler) Get(c fiber.Ctx) error {
	ctx := c.Context()
	actor := auth.ActorOf(c)

	var tz string
	if err := h.pool.QueryRow(ctx, `SELECT timezone FROM institution_settings LIMIT 1`).Scan(&tz); err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	today := attendance.MustLocalDate(time.Now(), tz)

	resp := dashboardResponse{Today: dashboardToday{}}

	asn, err := activeAssignment(ctx, h.pool, actor.ID, today)
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	resp.Assignment = asn

	attID, status, timeIn, timeOut, err := todayAttendance(ctx, h.pool, actor.ID, today)
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	resp.Today.AttendanceID = attID
	resp.Today.AttendanceStatus = status
	resp.Today.TimeInAt = timeIn
	resp.Today.TimeOutAt = timeOut

	if attID != nil {
		jID, jStatus, err := journalFor(ctx, h.pool, *attID)
		if err != nil {
			return httpx.Fail(c, httpx.Internal(err))
		}
		resp.Today.JournalID = jID
		resp.Today.JournalStatus = jStatus
	}

	if err := h.pool.QueryRow(ctx,
		`SELECT count(*) FROM notifications WHERE recipient_user_id = $1 AND read_at IS NULL`,
		actor.ID).Scan(&resp.UnreadNotifications); err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}

	resp.Today.NextAction = nextAction(resp)
	return httpx.OK(c, resp)
}

func nextAction(r dashboardResponse) string {
	switch {
	case r.Assignment == nil:
		return "contact_coordinator"
	case r.Today.AttendanceID == nil:
		return "time_in"
	case r.Today.AttendanceStatus != nil && *r.Today.AttendanceStatus == "open":
		return "time_out"
	case r.Today.JournalStatus == nil:
		return "view_summary" // attendance closed before journal service exists
	case *r.Today.JournalStatus == "needs_revision":
		return "revise_journal"
	case *r.Today.JournalStatus == "draft":
		return "complete_journal"
	default:
		return "view_summary"
	}
}

func activeAssignment(ctx context.Context, pool *db.Pool, userID uuid.UUID, today string) (*dashboardAssignment, error) {
	var a dashboardAssignment
	err := pool.QueryRow(ctx, `
		SELECT a.id, s.id, s.name, a.required_minutes,
		       COALESCE((
		         SELECT sum(s2.credited_minutes)::int
		         FROM attendance_sessions s2
		         WHERE s2.assignment_id = a.id AND s2.credited_minutes IS NOT NULL
		       ), 0) AS completed
		FROM ojt_assignments a
		JOIN trainee_profiles tp ON tp.id = a.trainee_id
		JOIN ojt_sites s ON s.id = a.site_id
		WHERE tp.user_id = $1
		  AND a.status = 'active'
		  AND a.start_date <= $2::date
		  AND (a.end_date IS NULL OR a.end_date >= $2::date)
		ORDER BY a.start_date DESC
		LIMIT 1`, userID, today).
		Scan(&a.ID, &a.Site.ID, &a.Site.Name, &a.RequiredMinutes, &a.CompletedMinutes)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	a.RemainingMinutes = max(a.RequiredMinutes-a.CompletedMinutes, 0)
	if a.RequiredMinutes > 0 {
		a.ProgressPercent = min(a.CompletedMinutes*100/a.RequiredMinutes, 100)
	}
	return &a, nil
}

func todayAttendance(ctx context.Context, pool *db.Pool, userID uuid.UUID, today string) (*uuid.UUID, *string, *time.Time, *time.Time, error) {
	var (
		id      uuid.UUID
		status  string
		timeIn  *time.Time
		timeOut *time.Time
	)
	err := pool.QueryRow(ctx, `
		SELECT s.id, s.status, s.effective_time_in_at, s.effective_time_out_at
		FROM attendance_sessions s
		JOIN trainee_profiles tp ON tp.id = s.trainee_id
		WHERE tp.user_id = $1 AND s.attendance_date = $2::date`, userID, today).
		Scan(&id, &status, &timeIn, &timeOut)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, nil, nil, nil
	}
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return &id, &status, timeIn, timeOut, nil
}

func journalFor(ctx context.Context, pool *db.Pool, attendanceID uuid.UUID) (*uuid.UUID, *string, error) {
	var (
		id     uuid.UUID
		status string
	)
	err := pool.QueryRow(ctx,
		`SELECT id, status FROM daily_journals WHERE attendance_id = $1`, attendanceID).
		Scan(&id, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	return &id, &status, nil
}
