// Staff reports (T124–T130). Reports derive from normalized
// attendance/journal data — no separate reporting ledger. Scope is always
// server-derived from the acting staff user.
package reports

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/skycode/ojt-management/backend/internal/platform/db"
	"github.com/skycode/ojt-management/backend/internal/platform/httpx"
)

// Column describes one report column for JSON consumers and CSV export.
type Column struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

// Report is a tabular result: columns + rows keyed by column key.
type Report struct {
	Name    string           `json:"name"`
	Columns []Column         `json:"columns"`
	Rows    []map[string]any `json:"rows"`
}

// Filters are the shared report query parameters.
type Filters struct {
	From, To  string
	TraineeID string
	SiteID    string
	Status    string
}

// reportDef maps a report name to its builder.
var reportDef = map[string]func(ctx context.Context, pool *db.Pool, scope []uuid.UUID, f Filters) (*Report, error){
	"daily-attendance":   dailyAttendance,
	"weekly-summary":     weeklySummary,
	"student-progress":   studentProgress,
	"journal-completion": journalCompletion,
	"exceptions":         exceptions,
	"completion":         completion,
}

// Names lists the supported report slugs (stable API surface).
func Names() []string {
	return []string{"daily-attendance", "weekly-summary", "student-progress",
		"journal-completion", "exceptions", "completion"}
}

// Run executes a named report with scope + filters.
func Run(ctx context.Context, pool *db.Pool, name string, scope []uuid.UUID, f Filters) (*Report, error) {
	fn, ok := reportDef[name]
	if !ok {
		return nil, httpx.ErrNotFound
	}
	r, err := fn(ctx, pool, scope, f)
	if err != nil {
		return nil, err
	}
	r.Name = name
	return r, nil
}

// scopeFilter is the authorized-trainee predicate on attendance_sessions
// (always arg $1). Assignment-based reports use scopeFilterT on
// trainee_profiles.id.
const scopeFilter = `($1::uuid[] IS NULL OR s.trainee_id = ANY($1))`
const scopeFilterT = `($1::uuid[] IS NULL OR t.id = ANY($1))`

func args(scope []uuid.UUID, f Filters) []any {
	var scopeArg any
	if scope != nil {
		scopeArg = scope
	}
	return []any{scopeArg, orNil(f.From), orNil(f.To),
		uuidOrNil(f.TraineeID), uuidOrNil(f.SiteID), orNil(f.Status)}
}

// shared filter predicates applied to every report query.
const pred = `
	AND ($2::date IS NULL OR s.attendance_date >= $2)
	AND ($3::date IS NULL OR s.attendance_date <= $3)
	AND ($4::uuid IS NULL OR s.trainee_id = $4)
	AND ($5::uuid IS NULL OR a.site_id = $5)
	AND ($6::text IS NULL OR s.status = $6)`

const joinBase = `
	FROM attendance_sessions s
	JOIN trainee_profiles t ON t.id = s.trainee_id
	JOIN users u ON u.id = t.user_id
	JOIN ojt_assignments a ON a.id = s.assignment_id
	JOIN ojt_sites si ON si.id = a.site_id
	LEFT JOIN daily_journals j ON j.attendance_id = s.id`

func rowsToMaps(rows pgx.Rows, cols []Column) ([]map[string]any, error) {
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			return nil, err
		}
		m := map[string]any{}
		for i, c := range cols {
			m[c.Key] = vals[i]
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func dailyAttendance(ctx context.Context, pool *db.Pool, scope []uuid.UUID, f Filters) (*Report, error) {
	cols := []Column{
		{Key: "date", Label: "Date"}, {Key: "trainee", Label: "Trainee"},
		{Key: "student_number", Label: "Student #"}, {Key: "site", Label: "Site"},
		{Key: "status", Label: "Status"}, {Key: "time_in", Label: "Time In"},
		{Key: "time_out", Label: "Time Out"}, {Key: "credited_minutes", Label: "Credited Minutes"},
		{Key: "flags", Label: "Flags"}, {Key: "journal_status", Label: "Journal"},
	}
	rows, err := pool.Query(ctx, `
		SELECT s.attendance_date::text, u.display_name, t.student_number, si.name,
		       s.status, s.effective_time_in_at, s.effective_time_out_at,
		       s.credited_minutes, COALESCE(array_to_string(s.flag_codes,','),''),
		       COALESCE(j.status,'')
		`+joinBase+` WHERE `+scopeFilter+pred+`
		ORDER BY s.attendance_date DESC, u.display_name`,
		args(scope, f)...)
	if err != nil {
		return nil, err
	}
	maps, err := rowsToMaps(rows, cols)
	return &Report{Columns: cols, Rows: maps}, err
}

func weeklySummary(ctx context.Context, pool *db.Pool, scope []uuid.UUID, f Filters) (*Report, error) {
	cols := []Column{
		{Key: "week_start", Label: "Week Start"}, {Key: "trainee", Label: "Trainee"},
		{Key: "student_number", Label: "Student #"}, {Key: "site", Label: "Site"},
		{Key: "days_present", Label: "Days Present"},
		{Key: "credited_minutes", Label: "Credited Minutes"},
		{Key: "journals_submitted", Label: "Journals Submitted"},
	}
	rows, err := pool.Query(ctx, `
		SELECT date_trunc('week', s.attendance_date)::date::text AS week_start,
		       u.display_name, t.student_number, si.name,
		       count(*) FILTER (WHERE s.status <> 'open'),
		       COALESCE(sum(s.credited_minutes),0),
		       count(*) FILTER (WHERE j.status IN ('submitted','reviewed'))
		`+joinBase+` WHERE `+scopeFilter+pred+`
		GROUP BY week_start, u.display_name, t.student_number, si.name
		ORDER BY week_start DESC, u.display_name`,
		args(scope, f)...)
	if err != nil {
		return nil, err
	}
	maps, err := rowsToMaps(rows, cols)
	return &Report{Columns: cols, Rows: maps}, err
}

func studentProgress(ctx context.Context, pool *db.Pool, scope []uuid.UUID, f Filters) (*Report, error) {
	cols := []Column{
		{Key: "trainee", Label: "Trainee"}, {Key: "student_number", Label: "Student #"},
		{Key: "site", Label: "Site"}, {Key: "required_minutes", Label: "Required Minutes"},
		{Key: "completed_minutes", Label: "Completed Minutes"},
		{Key: "progress_percent", Label: "Progress %"},
		{Key: "remaining_minutes", Label: "Remaining Minutes"},
		{Key: "status", Label: "Assignment Status"},
	}
	rows, err := pool.Query(ctx, `
		SELECT u.display_name, t.student_number, si.name,
		       a.required_minutes,
		       COALESCE(sum(s.credited_minutes) FILTER (WHERE s.status IN ('valid','flagged','corrected')),0),
		       a.status
		FROM ojt_assignments a
		JOIN trainee_profiles t ON t.id = a.trainee_id
		JOIN users u ON u.id = t.user_id
		JOIN ojt_sites si ON si.id = a.site_id
		LEFT JOIN attendance_sessions s ON s.assignment_id = a.id
		WHERE ($1::uuid[] IS NULL OR t.id = ANY($1))
		  AND ($4::uuid IS NULL OR t.id = $4)
		  AND ($5::uuid IS NULL OR a.site_id = $5)
		  AND ($6::text IS NULL OR a.status = $6)
		GROUP BY u.display_name, t.student_number, si.name, a.required_minutes, a.status
		ORDER BY u.display_name`,
		args(scope, f)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var name, sn, site, status string
		var req, comp int
		if err := rows.Scan(&name, &sn, &site, &req, &comp, &status); err != nil {
			return nil, err
		}
		pct := 0
		if req > 0 {
			pct = comp * 100 / req
			if pct > 100 {
				pct = 100
			}
		}
		out = append(out, map[string]any{
			"trainee": name, "student_number": sn, "site": site,
			"required_minutes": req, "completed_minutes": comp,
			"progress_percent": pct, "remaining_minutes": max(req-comp, 0),
			"status": status,
		})
	}
	return &Report{Columns: cols, Rows: out}, rows.Err()
}

func journalCompletion(ctx context.Context, pool *db.Pool, scope []uuid.UUID, f Filters) (*Report, error) {
	cols := []Column{
		{Key: "date", Label: "Date"}, {Key: "trainee", Label: "Trainee"},
		{Key: "student_number", Label: "Student #"}, {Key: "site", Label: "Site"},
		{Key: "status", Label: "Journal Status"}, {Key: "submitted_at", Label: "Submitted"},
		{Key: "reviewed_at", Label: "Reviewed"}, {Key: "revision_count", Label: "Revisions"},
	}
	rows, err := pool.Query(ctx, `
		SELECT s.attendance_date::text, u.display_name, t.student_number, si.name,
		       j.status, j.submitted_at, j.reviewed_at, j.revision_count
		FROM daily_journals j
		JOIN attendance_sessions s ON s.id = j.attendance_id
		JOIN trainee_profiles t ON t.id = j.trainee_id
		JOIN users u ON u.id = t.user_id
		JOIN ojt_assignments a ON a.id = s.assignment_id
		JOIN ojt_sites si ON si.id = a.site_id
		WHERE ($1::uuid[] IS NULL OR j.trainee_id = ANY($1))
		  AND ($2::date IS NULL OR s.attendance_date >= $2)
		  AND ($3::date IS NULL OR s.attendance_date <= $3)
		  AND ($4::uuid IS NULL OR j.trainee_id = $4)
		  AND ($5::uuid IS NULL OR a.site_id = $5)
		  AND ($6::text IS NULL OR j.status = $6)
		ORDER BY s.attendance_date DESC, u.display_name`,
		args(scope, f)...)
	if err != nil {
		return nil, err
	}
	maps, err := rowsToMaps(rows, cols)
	return &Report{Columns: cols, Rows: maps}, err
}

func exceptions(ctx context.Context, pool *db.Pool, scope []uuid.UUID, f Filters) (*Report, error) {
	cols := []Column{
		{Key: "date", Label: "Date"}, {Key: "trainee", Label: "Trainee"},
		{Key: "site", Label: "Site"}, {Key: "kind", Label: "Kind"},
		{Key: "detail", Label: "Detail"}, {Key: "status", Label: "Status"},
	}
	// Flags + pending/approved corrections, unioned into one exception feed.
	rows, err := pool.Query(ctx, `
		SELECT s.attendance_date::text, u.display_name, si.name,
		       'flag'::text, array_to_string(s.flag_codes,', '), s.status
		`+joinBase+`
		WHERE `+scopeFilter+pred+` AND cardinality(s.flag_codes) > 0
		UNION ALL
		SELECT s.attendance_date::text, u.display_name, si.name,
		       'correction'::text, c.type || ': ' || c.reason, c.status
		FROM correction_requests c
		JOIN attendance_sessions s ON s.id = c.attendance_id
		JOIN trainee_profiles t ON t.id = c.trainee_id
		JOIN users u ON u.id = t.user_id
		JOIN ojt_assignments a ON a.id = s.assignment_id
		JOIN ojt_sites si ON si.id = a.site_id
		WHERE `+scopeFilter+`
		  AND ($2::date IS NULL OR s.attendance_date >= $2)
		  AND ($3::date IS NULL OR s.attendance_date <= $3)
		  AND ($4::uuid IS NULL OR c.trainee_id = $4)
		  AND ($5::uuid IS NULL OR a.site_id = $5)
		  AND ($6::text IS NULL OR c.status = $6)
		ORDER BY 1 DESC, 2`,
		args(scope, f)...)
	if err != nil {
		return nil, err
	}
	maps, err := rowsToMaps(rows, cols)
	return &Report{Columns: cols, Rows: maps}, err
}

func completion(ctx context.Context, pool *db.Pool, scope []uuid.UUID, f Filters) (*Report, error) {
	cols := []Column{
		{Key: "trainee", Label: "Trainee"}, {Key: "student_number", Label: "Student #"},
		{Key: "site", Label: "Site"}, {Key: "required_minutes", Label: "Required Minutes"},
		{Key: "completed_minutes", Label: "Completed Minutes"},
		{Key: "progress_percent", Label: "Progress %"},
		{Key: "journals_reviewed", Label: "Journals Reviewed"},
		{Key: "completion_state", Label: "Completion"},
	}
	rows, err := pool.Query(ctx, `
		SELECT u.display_name, t.student_number, si.name, a.required_minutes,
		       COALESCE(sum(s.credited_minutes) FILTER (WHERE s.status IN ('valid','flagged','corrected')),0),
		       count(j.id) FILTER (WHERE j.status = 'reviewed'),
		       a.status
		FROM ojt_assignments a
		JOIN trainee_profiles t ON t.id = a.trainee_id
		JOIN users u ON u.id = t.user_id
		JOIN ojt_sites si ON si.id = a.site_id
		LEFT JOIN attendance_sessions s ON s.assignment_id = a.id
		LEFT JOIN daily_journals j ON j.attendance_id = s.id
		WHERE ($1::uuid[] IS NULL OR t.id = ANY($1))
		  AND ($4::uuid IS NULL OR t.id = $4)
		  AND ($5::uuid IS NULL OR a.site_id = $5)
		  AND ($6::text IS NULL OR a.status = $6)
		GROUP BY u.display_name, t.student_number, si.name, a.required_minutes, a.status
		ORDER BY u.display_name`,
		args(scope, f)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var name, sn, site, status string
		var req, comp, journalsReviewed int
		if err := rows.Scan(&name, &sn, &site, &req, &comp, &journalsReviewed, &status); err != nil {
			return nil, err
		}
		pct := 0
		if req > 0 {
			pct = comp * 100 / req
			if pct > 100 {
				pct = 100
			}
		}
		state := "in_progress"
		if pct >= 100 {
			state = "hours_complete"
		}
		out = append(out, map[string]any{
			"trainee": name, "student_number": sn, "site": site,
			"required_minutes": req, "completed_minutes": comp,
			"progress_percent": pct, "journals_reviewed": journalsReviewed,
			"completion_state": state,
		})
	}
	return &Report{Columns: cols, Rows: out}, rows.Err()
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
