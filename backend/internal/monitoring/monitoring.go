// Staff dashboard metrics + attendance monitoring (T109–T114).
// Every query is scope-filtered: coordinators see only their trainees;
// admins pass a nil scope for institution-wide visibility.
package monitoring

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/skycode/ojt-management/backend/internal/attendance"
	"github.com/skycode/ojt-management/backend/internal/platform/db"
	"github.com/skycode/ojt-management/backend/internal/platform/httpx"
)

type Metrics struct {
	ActiveTrainees         int  `json:"active_trainees"`
	PresentToday           int  `json:"present_today"`
	NoAttendance           int  `json:"no_attendance"`
	NoAttendanceConfigured bool `json:"no_attendance_configured"`
	FlaggedAttendance      int  `json:"flagged_attendance"`
	MissingJournals        int  `json:"missing_journals"`
	NearCompletion         int  `json:"near_completion"`
	CompletedHours         int  `json:"completed_hours"`
}

type CorrectionQueueItem struct {
	ID          uuid.UUID `json:"id"`
	TraineeName string    `json:"trainee_name"`
	Type        string    `json:"type"`
	Date        string    `json:"date"`
	RequestedAt time.Time `json:"requested_at"`
}

type FlagQueueItem struct {
	ID          uuid.UUID `json:"id"`
	TraineeName string    `json:"trainee_name"`
	Date        string    `json:"date"`
	Flags       []string  `json:"flags"`
}

type JournalQueueItem struct {
	ID          uuid.UUID  `json:"id"`
	TraineeName string     `json:"trainee_name"`
	Date        string     `json:"date"`
	SubmittedAt *time.Time `json:"submitted_at"`
}

type Queues struct {
	PendingCorrections       []CorrectionQueueItem `json:"pending_corrections"`
	RecentFlags              []FlagQueueItem       `json:"recent_flags"`
	JournalsNeedingAttention []JournalQueueItem    `json:"journals_needing_attention"`
}

type DashboardData struct {
	Metrics Metrics `json:"metrics"`
	Queues  Queues  `json:"queues"`
}

// scopeOn is the shared authorized-trainee filter (T109): nil scope arg
// means admin (institution-wide); a non-nil slice restricts to those trainees.
func scopeOn(alias string) string {
	return `($1::uuid[] IS NULL OR ` + alias + `.trainee_id = ANY($1))`
}

func scopeArg(scope []uuid.UUID) any {
	if scope == nil {
		return nil
	}
	return scope
}

// Dashboard computes scope-aware metrics and attention queues.
func Dashboard(ctx context.Context, pool *db.Pool, scope []uuid.UUID, now time.Time) (*DashboardData, error) {
	var tz string
	var journalCutoffH, unusualMin int
	var nearPct float64
	if err := pool.QueryRow(ctx, `
		SELECT timezone, missing_journal_cutoff_hours, unusual_session_minutes, near_completion_percent
		FROM institution_settings LIMIT 1`).
		Scan(&tz, &journalCutoffH, &unusualMin, &nearPct); err != nil {
		return nil, err
	}
	today, err := attendance.LocalDate(now, tz)
	if err != nil {
		return nil, err
	}
	loc, _ := time.LoadLocation(tz)
	weekday := int(now.In(loc).Weekday())
	if weekday == 0 {
		weekday = 7
	}

	var m Metrics
	s := scopeArg(scope)

	if err := pool.QueryRow(ctx, `
		SELECT count(DISTINCT a.trainee_id) FROM ojt_assignments a
		WHERE a.status = 'active' AND `+scopeOn("a"), s).Scan(&m.ActiveTrainees); err != nil {
		return nil, err
	}
	if err := pool.QueryRow(ctx, `
		SELECT count(DISTINCT s2.trainee_id) FROM attendance_sessions s2
		WHERE s2.attendance_date = $2 AND `+scopeOn("s2"), s, today).Scan(&m.PresentToday); err != nil {
		return nil, err
	}
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM attendance_sessions s2
		WHERE s2.attendance_date = $2 AND s2.status = 'flagged' AND `+scopeOn("s2"),
		s, today).Scan(&m.FlaggedAttendance); err != nil {
		return nil, err
	}
	// No Attendance: active assignments expecting today with no session.
	if err := pool.QueryRow(ctx, `
		SELECT count(*), EXISTS(
		    SELECT 1 FROM ojt_assignments a2
		    WHERE a2.status = 'active' AND cardinality(a2.expected_weekdays) > 0
		      AND ($1::uuid[] IS NULL OR a2.trainee_id = ANY($1)))
		FROM ojt_assignments a
		WHERE a.status = 'active'
		  AND cardinality(a.expected_weekdays) > 0
		  AND $3 = ANY(a.expected_weekdays)
		  AND a.start_date <= $2 AND (a.end_date IS NULL OR a.end_date >= $2)
		  AND `+scopeOn("a")+`
		  AND NOT EXISTS (SELECT 1 FROM attendance_sessions s2
		      WHERE s2.trainee_id = a.trainee_id AND s2.attendance_date = $2)`,
		s, today, weekday).Scan(&m.NoAttendance, &m.NoAttendanceConfigured); err != nil {
		return nil, err
	}
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM daily_journals j
		JOIN attendance_sessions s2 ON s2.id = j.attendance_id
		WHERE j.status IN ('draft','needs_revision')
		  AND COALESCE(s2.closed_at, s2.effective_time_out_at) < $2
		  AND ($1::uuid[] IS NULL OR j.trainee_id = ANY($1))`,
		s, now.Add(-time.Duration(journalCutoffH)*time.Hour)).Scan(&m.MissingJournals); err != nil {
		return nil, err
	}
	// Completion metrics over active assignments: credited minutes per
	// assignment vs required_minutes (same definition as progress).
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FILTER (WHERE pct >= $2 AND pct < 100),
		       count(*) FILTER (WHERE pct >= 100)
		FROM (
		    SELECT a.id,
		           COALESCE(sum(s2.credited_minutes),0)::float / NULLIF(a.required_minutes,0) * 100 AS pct
		    FROM ojt_assignments a
		    LEFT JOIN attendance_sessions s2
		        ON s2.assignment_id = a.id AND s2.status IN ('valid','flagged','corrected')
		    WHERE a.status = 'active' AND `+scopeOn("a")+`
		    GROUP BY a.id, a.required_minutes) t`,
		s, nearPct).Scan(&m.NearCompletion, &m.CompletedHours); err != nil {
		return nil, err
	}

	d := &DashboardData{Metrics: m}
	if d.Queues.PendingCorrections, err = pendingCorrections(ctx, pool, s); err != nil {
		return nil, err
	}
	if d.Queues.RecentFlags, err = recentFlags(ctx, pool, s); err != nil {
		return nil, err
	}
	if d.Queues.JournalsNeedingAttention, err = journalsAttention(ctx, pool, s); err != nil {
		return nil, err
	}
	return d, nil
}

func pendingCorrections(ctx context.Context, pool *db.Pool, scope any) ([]CorrectionQueueItem, error) {
	rows, err := pool.Query(ctx, `
		SELECT c.id, u.display_name, c.type, s.attendance_date::text, c.requested_at
		FROM correction_requests c
		JOIN attendance_sessions s ON s.id = c.attendance_id
		JOIN trainee_profiles tp ON tp.id = c.trainee_id
		JOIN users u ON u.id = tp.user_id
		WHERE c.status = 'pending' AND ($1::uuid[] IS NULL OR c.trainee_id = ANY($1))
		ORDER BY c.requested_at LIMIT 10`, scope)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []CorrectionQueueItem{}
	for rows.Next() {
		var it CorrectionQueueItem
		if err := rows.Scan(&it.ID, &it.TraineeName, &it.Type, &it.Date, &it.RequestedAt); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func recentFlags(ctx context.Context, pool *db.Pool, scope any) ([]FlagQueueItem, error) {
	rows, err := pool.Query(ctx, `
		SELECT s.id, u.display_name, s.attendance_date::text, s.flag_codes
		FROM attendance_sessions s
		JOIN trainee_profiles tp ON tp.id = s.trainee_id
		JOIN users u ON u.id = tp.user_id
		WHERE cardinality(s.flag_codes) > 0 AND ($1::uuid[] IS NULL OR s.trainee_id = ANY($1))
		ORDER BY s.attendance_date DESC LIMIT 10`, scope)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []FlagQueueItem{}
	for rows.Next() {
		var it FlagQueueItem
		if err := rows.Scan(&it.ID, &it.TraineeName, &it.Date, &it.Flags); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func journalsAttention(ctx context.Context, pool *db.Pool, scope any) ([]JournalQueueItem, error) {
	rows, err := pool.Query(ctx, `
		SELECT j.id, u.display_name, s.attendance_date::text, j.submitted_at
		FROM daily_journals j
		JOIN attendance_sessions s ON s.id = j.attendance_id
		JOIN trainee_profiles tp ON tp.id = j.trainee_id
		JOIN users u ON u.id = tp.user_id
		WHERE j.status = 'submitted' AND ($1::uuid[] IS NULL OR j.trainee_id = ANY($1))
		ORDER BY j.submitted_at LIMIT 10`, scope)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []JournalQueueItem{}
	for rows.Next() {
		var it JournalQueueItem
		if err := rows.Scan(&it.ID, &it.TraineeName, &it.Date, &it.SubmittedAt); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

// ---- staff attendance monitoring list (T114) --------------------------------

type AttendanceRow struct {
	ID              uuid.UUID  `json:"id"`
	Date            string     `json:"date"`
	TraineeName     string     `json:"trainee_name"`
	StudentNumber   string     `json:"student_number"`
	SiteName        string     `json:"site_name"`
	Status          string     `json:"status"`
	TimeInAt        time.Time  `json:"time_in_at"`
	TimeOutAt       *time.Time `json:"time_out_at"`
	CreditedMinutes *int       `json:"credited_minutes"`
	Flags           []string   `json:"flags"`
	JournalStatus   *string    `json:"journal_status"`
}

// AttendanceList is the paginated, scope-filtered staff monitoring table.
func AttendanceList(ctx context.Context, pool *db.Pool, scope []uuid.UUID,
	f struct{ From, To, Date, SiteID, TraineeID, Status string }, p httpx.Page) ([]AttendanceRow, int64, error) {

	where := `WHERE ($1::uuid[] IS NULL OR s.trainee_id = ANY($1))
		AND ($2::date IS NULL OR s.attendance_date >= $2)
		AND ($3::date IS NULL OR s.attendance_date <= $3)
		AND ($4::date IS NULL OR s.attendance_date = $4)
		AND ($5::uuid IS NULL OR a.site_id = $5)
		AND ($6::uuid IS NULL OR s.trainee_id = $6)
		AND ($7::text IS NULL OR s.status = $7)`
	args := []any{scopeArg(scope),
		orNil(f.From), orNil(f.To), orNil(f.Date),
		uuidOrNil(f.SiteID), uuidOrNil(f.TraineeID), orNil(f.Status)}

	var total int64
	countQ := `SELECT count(*) FROM attendance_sessions s
		JOIN ojt_assignments a ON a.id = s.assignment_id ` + where
	if err := pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := pool.Query(ctx, `
		SELECT s.id, s.attendance_date::text, u.display_name, tp.student_number,
		       si.name, s.status, s.effective_time_in_at, s.effective_time_out_at,
		       s.credited_minutes, s.flag_codes, j.status
		FROM attendance_sessions s
		JOIN trainee_profiles tp ON tp.id = s.trainee_id
		JOIN users u ON u.id = tp.user_id
		JOIN ojt_assignments a ON a.id = s.assignment_id
		JOIN ojt_sites si ON si.id = a.site_id
		LEFT JOIN daily_journals j ON j.attendance_id = s.id
		`+where+`
		ORDER BY s.attendance_date DESC, u.display_name
		LIMIT $8 OFFSET $9`, append(args, p.Limit(), p.Offset())...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []AttendanceRow{}
	for rows.Next() {
		var r AttendanceRow
		if err := rows.Scan(&r.ID, &r.Date, &r.TraineeName, &r.StudentNumber,
			&r.SiteName, &r.Status, &r.TimeInAt, &r.TimeOutAt,
			&r.CreditedMinutes, &r.Flags, &r.JournalStatus); err != nil {
			return nil, 0, err
		}
		out = append(out, r)
	}
	return out, total, rows.Err()
}

func orNil(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func uuidOrNil(s string) any {
	if s == "" {
		return nil
	}
	if id, err := uuid.Parse(s); err == nil {
		return id
	}
	return nil
}
