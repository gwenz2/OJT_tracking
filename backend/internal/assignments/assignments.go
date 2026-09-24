// Package assignments implements OJT assignment CRUD. An assignment links a
// trainee to a site with dates, required minutes, and the break policy used
// for server-side credited-minute computation.
package assignments

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/skycode/ojt-management/backend/internal/platform/db"
	"github.com/skycode/ojt-management/backend/internal/platform/httpx"
)

// Date serializes as YYYY-MM-DD in JSON.
type Date struct{ time.Time }

func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.Format("2006-01-02") + `"`), nil
}

func (d *Date) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		d.Time = time.Time{}
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}

type Assignment struct {
	ID                uuid.UUID `json:"id"`
	TraineeID         uuid.UUID `json:"trainee_id"`
	SiteID            uuid.UUID `json:"site_id"`
	SiteName          string    `json:"site_name,omitempty"`
	TraineeName       string    `json:"trainee_name,omitempty"`
	StartDate         Date      `json:"start_date"`
	EndDate           *Date     `json:"end_date"`
	RequiredMinutes   int       `json:"required_minutes"`
	BreakRuleType     string    `json:"break_rule_type"`
	BreakThresholdMin *int      `json:"break_threshold_minutes"`
	BreakDeductionMin int       `json:"break_deduction_minutes"`
	ExpectedWeekdays  []int     `json:"expected_weekdays"`
	Status            string    `json:"status"`
	CompletedMinutes  int       `json:"completed_minutes"`
	CreatedAt         time.Time `json:"created_at"`
}

// BreakRule is the contract's nested representation.
type BreakRule struct {
	Type             string `json:"type"`
	ThresholdMinutes *int   `json:"threshold_minutes"`
	DeductionMinutes *int   `json:"deduction_minutes"`
}

type CreateInput struct {
	TraineeID        string     `json:"trainee_id"`
	SiteID           string     `json:"site_id"`
	StartDate        string     `json:"start_date"`
	EndDate          *string    `json:"end_date"`
	RequiredMinutes  *int       `json:"required_minutes"`
	BreakRule        *BreakRule `json:"break_rule"`
	ExpectedWeekdays []int      `json:"expected_weekdays"`
	Status           string     `json:"status"`
}

type PatchInput struct {
	SiteID           *string    `json:"site_id"`
	StartDate        *string    `json:"start_date"`
	EndDate          *string    `json:"end_date"`
	RequiredMinutes  *int       `json:"required_minutes"`
	BreakRule        *BreakRule `json:"break_rule"`
	ExpectedWeekdays *[]int     `json:"expected_weekdays"`
	Status           *string    `json:"status"`
}

func toSmallint(days []int) []int16 {
	if days == nil {
		return nil
	}
	out := make([]int16, len(days))
	for i, d := range days {
		out[i] = int16(d)
	}
	return out
}

var validStatuses = map[string]bool{
	"planned": true, "active": true, "completed": true, "suspended": true, "cancelled": true,
}

func parseDate(s string) (time.Time, bool) {
	t, err := time.Parse("2006-01-02", strings.TrimSpace(s))
	return t, err == nil
}

func validateBreakRule(br *BreakRule, fields map[string]string) (ruleType string, threshold *int, deduction int) {
	ruleType, deduction = "none", 0
	if br == nil {
		return
	}
	switch br.Type {
	case "none":
	case "fixed_after_threshold":
		if br.ThresholdMinutes == nil || *br.ThresholdMinutes < 0 {
			fields["break_rule.threshold_minutes"] = "Threshold minutes must be zero or greater."
		} else {
			threshold = br.ThresholdMinutes
		}
		if br.DeductionMinutes == nil || *br.DeductionMinutes < 0 {
			fields["break_rule.deduction_minutes"] = "Deduction minutes must be zero or greater."
		} else {
			deduction = *br.DeductionMinutes
		}
		ruleType = "fixed_after_threshold"
	default:
		fields["break_rule.type"] = "Must be 'none' or 'fixed_after_threshold'."
	}
	return
}

func validateWeekdays(days []int, fields map[string]string) {
	for _, d := range days {
		if d < 1 || d > 7 {
			fields["expected_weekdays"] = "Weekdays must be 1 (Mon) through 7 (Sun)."
			return
		}
	}
}

const selectCols = `
	SELECT a.id, a.trainee_id, a.site_id, s.name, u.display_name, a.start_date, a.end_date,
	       a.required_minutes, a.break_rule_type, a.break_threshold_minutes,
	       a.break_deduction_minutes, a.expected_weekdays, a.status, a.created_at,
	       COALESCE((SELECT sum(att.credited_minutes) FROM attendance_sessions att
	                 WHERE att.assignment_id = a.id AND att.status IN ('valid','flagged','corrected')), 0)
	FROM ojt_assignments a
	JOIN ojt_sites s ON s.id = a.site_id
	JOIN trainee_profiles tp ON tp.id = a.trainee_id
	JOIN users u ON u.id = tp.user_id `

func scanAssignment(row pgx.Row) (*Assignment, error) {
	var a Assignment
	var start time.Time
	var end *time.Time
	var weekdays []int16
	err := row.Scan(&a.ID, &a.TraineeID, &a.SiteID, &a.SiteName, &a.TraineeName,
		&start, &end, &a.RequiredMinutes, &a.BreakRuleType, &a.BreakThresholdMin,
		&a.BreakDeductionMin, &weekdays, &a.Status, &a.CreatedAt, &a.CompletedMinutes)
	if err != nil {
		return nil, err
	}
	a.StartDate = Date{start}
	if end != nil {
		d := Date{*end}
		a.EndDate = &d
	}
	for _, d := range weekdays {
		a.ExpectedWeekdays = append(a.ExpectedWeekdays, int(d))
	}
	if a.ExpectedWeekdays == nil {
		a.ExpectedWeekdays = []int{}
	}
	return &a, nil
}

// Get returns one assignment by ID (no scope check — callers enforce).
func Get(ctx context.Context, pool *db.Pool, id uuid.UUID) (*Assignment, error) {
	a, err := scanAssignment(pool.QueryRow(ctx, selectCols+` WHERE a.id = $1`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httpx.ErrNotFound
		}
		return nil, httpx.Internal(err)
	}
	return a, nil
}

// ActiveForTrainee returns the trainee's active assignment joined with its
// site, or nil when none exists / none is in effect today.
func ActiveForTrainee(ctx context.Context, pool *db.Pool, traineeID uuid.UUID) (*Assignment, error) {
	a, err := scanAssignment(pool.QueryRow(ctx, selectCols+`
		WHERE a.trainee_id = $1 AND a.status = 'active'
		  AND a.start_date <= CURRENT_DATE
		  AND (a.end_date IS NULL OR a.end_date >= CURRENT_DATE)
		ORDER BY a.start_date DESC LIMIT 1`, traineeID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return a, nil
}

// List returns scope-filtered paginated assignments. scopeIDs nil = admin
// (unfiltered); empty non-nil = coordinator with no scope (matches nothing).
func List(ctx context.Context, pool *db.Pool, p httpx.Page, scopeIDs []uuid.UUID, traineeID, siteID, status string) ([]Assignment, int64, error) {
	where := `WHERE ($1::uuid[] IS NULL OR a.trainee_id = ANY($1))
		AND ($2::uuid IS NULL OR a.trainee_id = $2)
		AND ($3::uuid IS NULL OR a.site_id = $3)
		AND ($4::text IS NULL OR a.status = $4)`
	var scopeArg, traineeArg, siteArg, statusArg any
	if scopeIDs != nil {
		scopeArg = scopeIDs
	}
	if v, err := uuid.Parse(traineeID); err == nil {
		traineeArg = v
	}
	if v, err := uuid.Parse(siteID); err == nil {
		siteArg = v
	}
	if validStatuses[status] {
		statusArg = status
	}

	var total int64
	err := pool.QueryRow(ctx,
		`SELECT count(*) FROM ojt_assignments a `+where, scopeArg, traineeArg, siteArg, statusArg).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := pool.Query(ctx, selectCols+where+`
		ORDER BY a.created_at DESC LIMIT $5 OFFSET $6`,
		scopeArg, traineeArg, siteArg, statusArg, p.Limit(), p.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []Assignment
	for rows.Next() {
		var a Assignment
		var start time.Time
		var end *time.Time
		var weekdays []int16
		if err := rows.Scan(&a.ID, &a.TraineeID, &a.SiteID, &a.SiteName, &a.TraineeName,
			&start, &end, &a.RequiredMinutes, &a.BreakRuleType, &a.BreakThresholdMin,
			&a.BreakDeductionMin, &weekdays, &a.Status, &a.CreatedAt, &a.CompletedMinutes); err != nil {
			return nil, 0, err
		}
		a.StartDate = Date{start}
		if end != nil {
			d := Date{*end}
			a.EndDate = &d
		}
		for _, d := range weekdays {
			a.ExpectedWeekdays = append(a.ExpectedWeekdays, int(d))
		}
		if a.ExpectedWeekdays == nil {
			a.ExpectedWeekdays = []int{}
		}
		out = append(out, a)
	}
	return out, total, rows.Err()
}

func Create(ctx context.Context, pool *db.Pool, in CreateInput, createdBy uuid.UUID) (*Assignment, error) {
	fields := map[string]string{}
	traineeID, err1 := uuid.Parse(strings.TrimSpace(in.TraineeID))
	if err1 != nil {
		fields["trainee_id"] = "A valid trainee is required."
	}
	siteID, err2 := uuid.Parse(strings.TrimSpace(in.SiteID))
	if err2 != nil {
		fields["site_id"] = "A valid site is required."
	}
	start, ok := parseDate(in.StartDate)
	if !ok {
		fields["start_date"] = "A valid start date (YYYY-MM-DD) is required."
	}
	var end *time.Time
	if in.EndDate != nil && strings.TrimSpace(*in.EndDate) != "" {
		if t, ok2 := parseDate(*in.EndDate); ok2 {
			end = &t
		} else {
			fields["end_date"] = "End date must be YYYY-MM-DD."
		}
	}
	if end != nil && ok && end.Before(start) {
		fields["end_date"] = "End date must not be before the start date."
	}
	if in.RequiredMinutes == nil || *in.RequiredMinutes <= 0 {
		fields["required_minutes"] = "Required minutes must be a positive integer."
	}
	status := "planned"
	if in.Status != "" {
		if !validStatuses[in.Status] {
			fields["status"] = "Invalid status."
		} else {
			status = in.Status
		}
	}
	validateWeekdays(in.ExpectedWeekdays, fields)
	ruleType, threshold, deduction := validateBreakRule(in.BreakRule, fields)
	if len(fields) > 0 {
		return nil, httpx.Validation(fields)
	}

	// Referenced rows must exist; site must be active for a new assignment.
	var siteActive bool
	if err := pool.QueryRow(ctx, `SELECT is_active FROM ojt_sites WHERE id = $1`, siteID).Scan(&siteActive); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httpx.Validation(map[string]string{"site_id": "Site not found."})
		}
		return nil, httpx.Internal(err)
	}
	if !siteActive {
		return nil, httpx.Validation(map[string]string{"site_id": "Site is inactive."})
	}
	var traineeExists bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM trainee_profiles WHERE id = $1)`, traineeID).Scan(&traineeExists); err != nil {
		return nil, httpx.Internal(err)
	}
	if !traineeExists {
		return nil, httpx.Validation(map[string]string{"trainee_id": "Trainee not found."})
	}

	var id uuid.UUID
	err := pool.QueryRow(ctx, `
		INSERT INTO ojt_assignments
			(trainee_id, site_id, start_date, end_date, required_minutes,
			 break_rule_type, break_threshold_minutes, break_deduction_minutes,
			 expected_weekdays, status, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id`,
		traineeID, siteID, start, end, *in.RequiredMinutes,
		ruleType, threshold, deduction, toSmallint(in.ExpectedWeekdays), status, createdBy).Scan(&id)
	if err != nil {
		if db.IsUniqueViolation(err, "ojt_assignments_one_active_per_trainee") {
			return nil, httpx.Conflict("ACTIVE_ASSIGNMENT_EXISTS",
				"This trainee already has an active assignment. Suspend or complete it first.")
		}
		return nil, httpx.Internal(err)
	}
	return Get(ctx, pool, id)
}

func Update(ctx context.Context, pool *db.Pool, id uuid.UUID, in PatchInput) (*Assignment, error) {
	cur, err := Get(ctx, pool, id)
	if err != nil {
		return nil, err
	}

	siteID := cur.SiteID
	if in.SiteID != nil {
		if v, err := uuid.Parse(*in.SiteID); err == nil {
			siteID = v
		} else {
			return nil, httpx.Validation(map[string]string{"site_id": "A valid site is required."})
		}
	}
	start := cur.StartDate.Time
	if in.StartDate != nil {
		if t, ok := parseDate(*in.StartDate); ok {
			start = t
		} else {
			return nil, httpx.Validation(map[string]string{"start_date": "Start date must be YYYY-MM-DD."})
		}
	}
	var end *time.Time
	if cur.EndDate != nil {
		e := cur.EndDate.Time
		end = &e
	}
	if in.EndDate != nil {
		if strings.TrimSpace(*in.EndDate) == "" {
			end = nil
		} else if t, ok := parseDate(*in.EndDate); ok {
			end = &t
		} else {
			return nil, httpx.Validation(map[string]string{"end_date": "End date must be YYYY-MM-DD."})
		}
	}
	if end != nil && end.Before(start) {
		return nil, httpx.Validation(map[string]string{"end_date": "End date must not be before the start date."})
	}
	required := cur.RequiredMinutes
	if in.RequiredMinutes != nil {
		if *in.RequiredMinutes <= 0 {
			return nil, httpx.Validation(map[string]string{"required_minutes": "Required minutes must be positive."})
		}
		required = *in.RequiredMinutes
	}
	status := cur.Status
	if in.Status != nil {
		if !validStatuses[*in.Status] {
			return nil, httpx.Validation(map[string]string{"status": "Invalid status."})
		}
		status = *in.Status
	}
	weekdays := cur.ExpectedWeekdays
	if in.ExpectedWeekdays != nil {
		fields := map[string]string{}
		validateWeekdays(*in.ExpectedWeekdays, fields)
		if len(fields) > 0 {
			return nil, httpx.Validation(fields)
		}
		weekdays = *in.ExpectedWeekdays
	}
	ruleType, threshold, deduction := cur.BreakRuleType, cur.BreakThresholdMin, cur.BreakDeductionMin
	if in.BreakRule != nil {
		fields := map[string]string{}
		ruleType, threshold, deduction = validateBreakRule(in.BreakRule, fields)
		if len(fields) > 0 {
			return nil, httpx.Validation(fields)
		}
	}

	_, err = pool.Exec(ctx, `
		UPDATE ojt_assignments SET
			site_id=$2, start_date=$3, end_date=$4, required_minutes=$5,
			break_rule_type=$6, break_threshold_minutes=$7, break_deduction_minutes=$8,
			expected_weekdays=$9, status=$10, updated_at=now()
		WHERE id=$1`,
		id, siteID, start, end, required, ruleType, threshold, deduction, toSmallint(weekdays), status)
	if err != nil {
		if db.IsUniqueViolation(err, "ojt_assignments_one_active_per_trainee") {
			return nil, httpx.Conflict("ACTIVE_ASSIGNMENT_EXISTS",
				"This trainee already has an active assignment.")
		}
		return nil, httpx.Internal(err)
	}
	return Get(ctx, pool, id)
}
