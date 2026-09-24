// Scheduled reminder generation (T120). Each reminder uses a deterministic
// event_key so repeated job runs never duplicate a logical notification.
//
//   - open_time_in: session still open past the unusual-session threshold
//   - missing_journal: closed day whose journal is still draft past the cutoff
//   - missing_attendance: expected weekday with no session (after cutoffHour local)
//
// Note: candidate IDs are collected and rows closed before inserting —
// pgx keeps the tx connection busy while a Rows object is open.
package reminders

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/skycode/ojt-management/backend/internal/attendance"
	"github.com/skycode/ojt-management/backend/internal/notifications"
	"github.com/skycode/ojt-management/backend/internal/platform/db"
)

// Run generates all due reminders inside one transaction. Returns counts per
// reminder kind actually inserted (deduped rows are skipped).
func Run(ctx context.Context, pool *db.Pool, now time.Time, missingAfterHour int) (openIn, missingJournal, missingAttendance int, err error) {
	var tz string
	var journalCutoff, unusualMin int
	if err := pool.QueryRow(ctx, `
		SELECT timezone, missing_journal_cutoff_hours, unusual_session_minutes
		FROM institution_settings LIMIT 1`).Scan(&tz, &journalCutoff, &unusualMin); err != nil {
		return 0, 0, 0, fmt.Errorf("settings: %w", err)
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("timezone %q: %w", tz, err)
	}
	today, err := attendance.LocalDate(now, tz)
	if err != nil {
		return 0, 0, 0, err
	}

	err = pool.InTx(ctx, func(tx pgx.Tx) error {
		var err error
		openIn, err = openTimeIn(ctx, tx, now, unusualMin)
		if err != nil {
			return err
		}
		missingJournal, err = missingJournals(ctx, tx, now, journalCutoff)
		if err != nil {
			return err
		}
		missingAttendance, err = missingAttendanceReminders(ctx, tx, now, today, loc, missingAfterHour)
		return err
	})
	return openIn, missingJournal, missingAttendance, err
}

// collect scans (resource_id, user_id) pairs and closes rows before returning.
func collect(ctx context.Context, rows pgx.Rows) ([][2]uuid.UUID, error) {
	defer rows.Close()
	var out [][2]uuid.UUID
	for rows.Next() {
		var pair [2]uuid.UUID
		if err := rows.Scan(&pair[0], &pair[1]); err != nil {
			return nil, err
		}
		out = append(out, pair)
	}
	return out, rows.Err()
}

// openTimeIn reminds trainees whose session has been open beyond the
// institution's unusual-session threshold.
func openTimeIn(ctx context.Context, tx pgx.Tx, now time.Time, unusualMin int) (int, error) {
	rows, err := tx.Query(ctx, `
		SELECT s.id, tp.user_id
		FROM attendance_sessions s
		JOIN trainee_profiles tp ON tp.id = s.trainee_id
		WHERE s.status = 'open'
		  AND s.effective_time_in_at < $1`, now.Add(-time.Duration(unusualMin)*time.Minute))
	if err != nil {
		return 0, err
	}
	pairs, err := collect(ctx, rows)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, p := range pairs {
		attID, userID := p[0], p[1]
		if ok, err := notifications.CreateTx(ctx, tx, userID,
			"reminder:open:"+attID.String(), "open_time_in",
			"Still timed in", "You have an open Time In. Remember to Time Out.",
			"attendance_session", &attID); err != nil {
			return n, err
		} else if ok {
			n++
		}
	}
	return n, nil
}

// missingJournals reminds trainees whose completed day lacks a submitted
// journal beyond the configured cutoff.
func missingJournals(ctx context.Context, tx pgx.Tx, now time.Time, cutoffHours int) (int, error) {
	rows, err := tx.Query(ctx, `
		SELECT j.id, tp.user_id
		FROM daily_journals j
		JOIN attendance_sessions s ON s.id = j.attendance_id
		JOIN trainee_profiles tp ON tp.id = j.trainee_id
		WHERE j.status IN ('draft','needs_revision')
		  AND COALESCE(s.closed_at, s.effective_time_out_at) < $1`,
		now.Add(-time.Duration(cutoffHours)*time.Hour))
	if err != nil {
		return 0, err
	}
	pairs, err := collect(ctx, rows)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, p := range pairs {
		jID, userID := p[0], p[1]
		if ok, err := notifications.CreateTx(ctx, tx, userID,
			"reminder:journal:"+jID.String(), "missing_journal",
			"Journal due", "Your daily journal for a completed day is still unfinished.",
			"daily_journal", &jID); err != nil {
			return n, err
		} else if ok {
			n++
		}
	}
	return n, nil
}

// missingAttendanceReminders reminds trainees with an active assignment whose
// expected weekdays include today but who have no attendance session — only
// once the configured local hour has passed.
func missingAttendanceReminders(ctx context.Context, tx pgx.Tx, now time.Time, today string, loc *time.Location, afterHour int) (int, error) {
	localNow := now.In(loc)
	if localNow.Hour() < afterHour {
		return 0, nil
	}
	weekday := int(localNow.Weekday()) // Go: Sunday=0
	if weekday == 0 {
		weekday = 7 // schema uses 1..7 (ISO)
	}
	rows, err := tx.Query(ctx, `
		SELECT a.id, tp.user_id
		FROM ojt_assignments a
		JOIN trainee_profiles tp ON tp.id = a.trainee_id
		WHERE a.status = 'active'
		  AND cardinality(a.expected_weekdays) > 0
		  AND $2 = ANY(a.expected_weekdays)
		  AND a.start_date <= $1
		  AND (a.end_date IS NULL OR a.end_date >= $1)
		  AND NOT EXISTS (
		      SELECT 1 FROM attendance_sessions s
		      WHERE s.trainee_id = a.trainee_id AND s.attendance_date = $1)`,
		today, weekday)
	if err != nil {
		return 0, err
	}
	pairs, err := collect(ctx, rows)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, p := range pairs {
		aID, userID := p[0], p[1]
		if ok, err := notifications.CreateTx(ctx, tx, userID,
			fmt.Sprintf("reminder:attendance:%s:%s", aID, today), "missing_attendance",
			"No attendance today", "You are expected on-site today and have not timed in.",
			"ojt_assignment", &aID); err != nil {
			return n, err
		} else if ok {
			n++
		}
	}
	return n, nil
}
