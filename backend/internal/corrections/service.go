// Correction requests: trainee proposals + staff decisions.
//
// Invariants (spec FR-040..045, data-model §3.12–3.13, §6 correction tx):
//   - one pending correction per attendance (partial unique index)
//   - approved corrections insert an attendance_adjustments row and update
//     only the effective_* projection — original_* evidence is untouched
//   - decision runs under FOR UPDATE locks on correction + attendance rows
package corrections

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/skycode/ojt-management/backend/internal/attendance"
	"github.com/skycode/ojt-management/backend/internal/audit"
	"github.com/skycode/ojt-management/backend/internal/notifications"
	"github.com/skycode/ojt-management/backend/internal/platform/db"
	"github.com/skycode/ojt-management/backend/internal/platform/httpx"
)

// allowedProposals maps correction type → which proposed fields are permitted.
// missed_time_out requires a proposed time-out; incorrect_time and
// field_assignment need at least one proposed timestamp; gps_issue/other are
// review-driven and may carry no proposed times.
var allowedProposals = map[string]struct {
	needIn, needOut, any bool
}{
	"missed_time_out":  {needOut: true},
	"incorrect_time":   {any: true},
	"field_assignment": {any: true},
	"gps_issue":        {},
	"other":            {},
}

type Correction struct {
	ID                uuid.UUID  `json:"id"`
	AttendanceID      uuid.UUID  `json:"attendance_id"`
	AttendanceDate    string     `json:"attendance_date,omitempty"`
	Type              string     `json:"type"`
	Reason            string     `json:"reason"`
	ProposedTimeInAt  *time.Time `json:"proposed_time_in_at"`
	ProposedTimeOutAt *time.Time `json:"proposed_time_out_at"`
	Status            string     `json:"status"`
	RequestedAt       time.Time  `json:"requested_at"`
	DecidedAt         *time.Time `json:"decided_at"`
	DecisionComment   *string    `json:"decision_comment"`
	TraineeID         uuid.UUID  `json:"trainee_id,omitempty"`
	TraineeName       string     `json:"trainee_name,omitempty"`
	SiteName          string     `json:"site_name,omitempty"`
}

// RequestInput is a trainee correction proposal.
type RequestInput struct {
	AttendanceID      uuid.UUID
	Type              string
	Reason            string
	ProposedTimeInAt  *time.Time
	ProposedTimeOutAt *time.Time
}

// validateProposal enforces per-type field rules and sane timestamps.
func validateProposal(in RequestInput, effectiveIn time.Time, now time.Time) error {
	rule, ok := allowedProposals[in.Type]
	if !ok {
		return httpx.Validation(map[string]string{"type": "Unknown correction type."})
	}
	if strings.TrimSpace(in.Reason) == "" {
		return httpx.Validation(map[string]string{"reason": "A reason is required."})
	}
	bad := httpx.Unprocessable("INVALID_PROPOSED_TIME", "Proposed times are invalid.")
	if rule.needIn && in.ProposedTimeInAt == nil ||
		rule.needOut && in.ProposedTimeOutAt == nil ||
		rule.any && in.ProposedTimeInAt == nil && in.ProposedTimeOutAt == nil {
		return bad
	}
	if in.ProposedTimeInAt != nil && in.ProposedTimeInAt.After(now) {
		return bad
	}
	if in.ProposedTimeOutAt != nil && in.ProposedTimeOutAt.After(now) {
		return bad
	}
	effIn := effectiveIn
	if in.ProposedTimeInAt != nil {
		effIn = *in.ProposedTimeInAt
	}
	if in.ProposedTimeOutAt != nil && !in.ProposedTimeOutAt.After(effIn) {
		return bad
	}
	return nil
}

// Request creates a pending correction for an owned attendance session.
func Request(ctx context.Context, pool *db.Pool, traineeID, actorUserID uuid.UUID, in RequestInput) (*Correction, error) {
	// Ownership + current effective time-in for validation.
	var effectiveIn time.Time
	err := pool.QueryRow(ctx, `
		SELECT effective_time_in_at FROM attendance_sessions
		WHERE id = $1 AND trainee_id = $2`, in.AttendanceID, traineeID).Scan(&effectiveIn)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httpx.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := validateProposal(in, effectiveIn, time.Now()); err != nil {
		return nil, err
	}

	var c Correction
	err = pool.InTx(ctx, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, `
			INSERT INTO correction_requests
			    (attendance_id, trainee_id, type, reason, proposed_time_in_at, proposed_time_out_at)
			VALUES ($1,$2,$3,$4,$5,$6)
			RETURNING id, attendance_id, type, reason, proposed_time_in_at,
			          proposed_time_out_at, status, requested_at`,
			in.AttendanceID, traineeID, in.Type, in.Reason,
			in.ProposedTimeInAt, in.ProposedTimeOutAt).
			Scan(&c.ID, &c.AttendanceID, &c.Type, &c.Reason,
				&c.ProposedTimeInAt, &c.ProposedTimeOutAt, &c.Status, &c.RequestedAt)
		if err != nil {
			if strings.Contains(err.Error(), "correction_requests_one_pending") {
				return httpx.Conflict("CORRECTION_ALREADY_PENDING",
					"A correction request is already pending for this day.")
			}
			return err
		}
		audit.Record(ctx, tx, &actorUserID, "correction.requested",
			"correction_request", &c.ID, "", map[string]any{"type": in.Type})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// ListMine is the trainee's own request history.
func ListMine(ctx context.Context, pool *db.Pool, traineeID uuid.UUID, p httpx.Page) ([]Correction, int64, error) {
	var total int64
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM correction_requests WHERE trainee_id = $1`, traineeID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := pool.Query(ctx, `
		SELECT c.id, c.attendance_id, s.attendance_date::text, c.type, c.reason,
		       c.proposed_time_in_at, c.proposed_time_out_at, c.status,
		       c.requested_at, c.decided_at, c.decision_comment
		FROM correction_requests c
		JOIN attendance_sessions s ON s.id = c.attendance_id
		WHERE c.trainee_id = $1
		ORDER BY c.requested_at DESC LIMIT $2 OFFSET $3`, traineeID, p.Limit(), p.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []Correction{}
	for rows.Next() {
		var c Correction
		if err := rows.Scan(&c.ID, &c.AttendanceID, &c.AttendanceDate, &c.Type, &c.Reason,
			&c.ProposedTimeInAt, &c.ProposedTimeOutAt, &c.Status,
			&c.RequestedAt, &c.DecidedAt, &c.DecisionComment); err != nil {
			return nil, 0, err
		}
		out = append(out, c)
	}
	return out, total, rows.Err()
}

// ListStaff is the scope-filtered staff queue.
func ListStaff(ctx context.Context, pool *db.Pool, scope []uuid.UUID, status, cType, traineeID, from, to string, p httpx.Page) ([]Correction, int64, error) {
	where := `WHERE ($1::uuid[] IS NULL OR c.trainee_id = ANY($1))
		AND ($2::text IS NULL OR c.status = $2)
		AND ($3::text IS NULL OR c.type = $3)
		AND ($4::uuid IS NULL OR c.trainee_id = $4)
		AND ($5::date IS NULL OR s.attendance_date >= $5)
		AND ($6::date IS NULL OR s.attendance_date <= $6)`
	var scopeArg any
	if scope != nil {
		scopeArg = scope
	}
	args := []any{scopeArg, orNil(status), orNil(cType), orNilUUID(traineeID), orNil(from), orNil(to)}

	var total int64
	countQ := `SELECT count(*) FROM correction_requests c
		JOIN attendance_sessions s ON s.id = c.attendance_id ` + where
	if err := pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	q := `SELECT c.id, c.attendance_id, s.attendance_date::text, c.type, c.reason,
		c.proposed_time_in_at, c.proposed_time_out_at, c.status,
		c.requested_at, c.decided_at, c.decision_comment,
		c.trainee_id, u.display_name, si.name
	FROM correction_requests c
	JOIN attendance_sessions s ON s.id = c.attendance_id
	JOIN trainee_profiles tp ON tp.id = c.trainee_id
	JOIN users u ON u.id = tp.user_id
	JOIN ojt_assignments a ON a.id = s.assignment_id
	JOIN ojt_sites si ON si.id = a.site_id ` + where +
		` ORDER BY CASE c.status WHEN 'pending' THEN 0 ELSE 1 END, c.requested_at DESC
	 LIMIT $7 OFFSET $8`
	rows, err := pool.Query(ctx, q, append(args, p.Limit(), p.Offset())...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []Correction{}
	for rows.Next() {
		var c Correction
		if err := rows.Scan(&c.ID, &c.AttendanceID, &c.AttendanceDate, &c.Type, &c.Reason,
			&c.ProposedTimeInAt, &c.ProposedTimeOutAt, &c.Status,
			&c.RequestedAt, &c.DecidedAt, &c.DecisionComment,
			&c.TraineeID, &c.TraineeName, &c.SiteName); err != nil {
			return nil, 0, err
		}
		out = append(out, c)
	}
	return out, total, rows.Err()
}

// TraineeIDOf returns the owning trainee for scope checks.
func TraineeIDOf(ctx context.Context, pool *db.Pool, correctionID uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID
	err := pool.QueryRow(ctx,
		`SELECT trainee_id FROM correction_requests WHERE id = $1`, correctionID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, httpx.ErrNotFound
	}
	return id, err
}

// Decide applies an approve/reject decision under row locks.
// Approval writes an immutable adjustment and recomputes the effective
// projection + credited minutes; the original evidence is never touched.
func Decide(ctx context.Context, pool *db.Pool, correctionID, staffID uuid.UUID, decision string, comment *string) error {
	if decision != "approved" && decision != "rejected" {
		return httpx.Validation(map[string]string{"decision": "Must be approved or rejected."})
	}
	return pool.InTx(ctx, func(tx pgx.Tx) error {
		var status, attendanceID string
		var traineeID uuid.UUID
		var assignmentID uuid.UUID
		var pIn, pOut *time.Time
		err := tx.QueryRow(ctx, `
			SELECT status, attendance_id, trainee_id, proposed_time_in_at, proposed_time_out_at
			FROM correction_requests WHERE id = $1 FOR UPDATE`, correctionID).
			Scan(&status, &attendanceID, &traineeID, &pIn, &pOut)
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		}
		if err != nil {
			return err
		}
		if status != "pending" {
			return httpx.Conflict("CORRECTION_NOT_PENDING",
				"This correction request has already been decided.")
		}

		if decision == "rejected" {
			if _, err := tx.Exec(ctx, `
				UPDATE correction_requests SET status = 'rejected', decided_by = $2,
				    decision_comment = $3, decided_at = now(), updated_at = now()
				WHERE id = $1`, correctionID, staffID, comment); err != nil {
				return err
			}
		} else {
			// Lock the attendance row and recompute the effective projection.
			var curIn time.Time
			var curOut *time.Time
			var attStatus string
			err := tx.QueryRow(ctx, `
				SELECT effective_time_in_at, effective_time_out_at, assignment_id, status
				FROM attendance_sessions WHERE id = $1 FOR UPDATE`, attendanceID).
				Scan(&curIn, &curOut, &assignmentID, &attStatus)
			if err != nil {
				return err
			}
			newIn := curIn
			if pIn != nil {
				newIn = *pIn
			}
			newOut := curOut
			if pOut != nil {
				newOut = pOut
			}
			if newOut != nil && !newOut.After(newIn) {
				return httpx.Unprocessable("INVALID_EFFECTIVE_DURATION",
					"Effective time out must be after time in.")
			}

			// Break rule comes from the assignment — server-owned policy.
			var rule attendance.BreakRule
			if err := tx.QueryRow(ctx, `
				SELECT break_rule_type, COALESCE(break_threshold_minutes,0), break_deduction_minutes
				FROM ojt_assignments WHERE id = $1`, assignmentID).
				Scan(&rule.Type, &rule.ThresholdMinutes, &rule.DeductionMinutes); err != nil {
				return err
			}
			var credited *int
			newStatus := "corrected"
			if newOut != nil {
				_, _, c := attendance.CreditedMinutes(
					attendance.SessionTimes{In: newIn, Out: *newOut}, rule)
				credited = &c
				// Keep a flagged marker if the session was flagged.
				if attStatus == "flagged" {
					newStatus = "corrected"
				}
			}

			if _, err := tx.Exec(ctx, `
				INSERT INTO attendance_adjustments
				    (attendance_id, correction_request_id, effective_time_in_at,
				     effective_time_out_at, credited_minutes, approved_by)
				VALUES ($1,$2,$3,$4,$5,$6)`,
				attendanceID, correctionID, newIn, newOut, credited, staffID); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `
				UPDATE attendance_sessions SET effective_time_in_at = $2,
				    effective_time_out_at = $3, credited_minutes = COALESCE($4, credited_minutes),
				    status = $5, version = version + 1, updated_at = now()
				WHERE id = $1`, attendanceID, newIn, newOut, credited, newStatus); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `
				UPDATE correction_requests SET status = 'approved', decided_by = $2,
				    decision_comment = $3, decided_at = now(), updated_at = now()
				WHERE id = $1`, correctionID, staffID, comment); err != nil {
				return err
			}
			// A correction that closes an open day must yield its journal.
			if newOut != nil && curOut == nil {
				if _, err := tx.Exec(ctx, `
					UPDATE attendance_sessions SET closed_at = COALESCE(closed_at, now())
					WHERE id = $1`, attendanceID); err != nil {
					return err
				}
				if _, err := ensureJournal(ctx, tx, attendanceID, traineeID); err != nil {
					return err
				}
			}
		}

		// Notify the trainee (dedup by event key).
		var traineeUser uuid.UUID
		if err := tx.QueryRow(ctx,
			`SELECT user_id FROM trainee_profiles WHERE id = $1`, traineeID).Scan(&traineeUser); err != nil {
			return err
		}
		// An approved correction may tip the trainee over required hours.
		if decision == "approved" {
			if err := notifications.NotifyCompletionIfReachedTx(ctx, tx,
				traineeUser, traineeID, assignmentID); err != nil {
				return err
			}
		}
		title := "Correction request approved"
		if decision == "rejected" {
			title = "Correction request rejected"
		}
		if _, err := notifications.CreateTx(ctx, tx, traineeUser,
			fmt.Sprintf("correction:%s:%s", correctionID, decision),
			"correction_decided", title,
			"Your correction request was "+decidedVerb(decision)+".",
			"correction_request", &correctionID); err != nil {
			return err
		}

		audit.Record(ctx, tx, &staffID, "correction."+decision,
			"correction_request", &correctionID, "", nil)
		return nil
	})
}

func decidedVerb(d string) string {
	if d == "approved" {
		return "approved"
	}
	return "rejected"
}

func ensureJournal(ctx context.Context, tx pgx.Tx, attendanceID string, traineeID uuid.UUID) (uuid.UUID, error) {
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

func orNil(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func orNilUUID(s string) any {
	if s == "" {
		return nil
	}
	if id, err := uuid.Parse(s); err == nil {
		return id
	}
	return nil
}
