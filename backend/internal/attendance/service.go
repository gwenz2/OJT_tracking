// Attendance write service: Time In / Time Out.
//
// Invariants enforced here (spec §2):
//   - server timestamp is the official time; device time is metadata only
//   - one session per trainee per institution-local date (DB unique index +
//     this service's canonical-result replay on unique violations)
//   - idempotent retries via (trainee_id, client_action_id) unique index
//   - weak/outside/unavailable GPS is stored and flagged, never dropped
//   - credited minutes computed server-side in integer minutes
package attendance

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/skycode/ojt-management/backend/internal/audit"
	"github.com/skycode/ojt-management/backend/internal/notifications"
	"github.com/skycode/ojt-management/backend/internal/platform/db"
	"github.com/skycode/ojt-management/backend/internal/platform/httpx"
	"github.com/skycode/ojt-management/backend/internal/platform/storage"
)

type Service struct {
	pool     *db.Pool
	store    *storage.Store
	maxBytes int64
}

func NewService(pool *db.Pool, store *storage.Store, maxUploadBytes int64) *Service {
	return &Service{pool: pool, store: store, maxBytes: maxUploadBytes}
}

// ---- inputs ---------------------------------------------------------------

// EvidenceInput is the parsed multipart payload. Every field except
// ClientActionID and Photo is optional metadata.
type EvidenceInput struct {
	ClientActionID          uuid.UUID
	Photo                   []byte
	DeviceCapturedAt        *time.Time
	DeviceTZOffsetMin       *int
	Latitude                *float64
	Longitude               *float64
	GPSAccuracyM            *float64
	LocationExceptionReason string
}

// ---- outputs --------------------------------------------------------------

type AttendanceView struct {
	ID                uuid.UUID  `json:"id"`
	Date              string     `json:"date"`
	Status            string     `json:"status"`
	OfficialTimeInAt  time.Time  `json:"official_time_in_at"`
	OfficialTimeOutAt *time.Time `json:"official_time_out_at"`
	CreditedMinutes   *int       `json:"credited_minutes,omitempty"`
	LocationStatus    string     `json:"location_status,omitempty"`
	DistanceFromSiteM *float64   `json:"distance_from_site_m,omitempty"`
	GPSAccuracyM      *float64   `json:"gps_accuracy_m,omitempty"`
	Flags             []string   `json:"flags"`
}

type ProgressView struct {
	CompletedMinutes int `json:"completed_minutes"`
	RequiredMinutes  int `json:"required_minutes"`
	RemainingMinutes int `json:"remaining_minutes"`
	ProgressPercent  int `json:"progress_percent"`
}

type TimeInResult struct {
	Attendance AttendanceView `json:"attendance"`
}

type TimeOutResult struct {
	Attendance AttendanceView `json:"attendance"`
	Progress   ProgressView   `json:"progress"`
	Journal    *journalRef    `json:"journal"`
}

type journalRef struct {
	ID     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}

// ---- shared context -------------------------------------------------------

type traineeCtx struct {
	TraineeID     uuid.UUID
	DisplayName   string
	StudentNumber string
	Timezone      string
	LowAccuracyM  float64
	UnusualMin    int
}

func (s *Service) traineeCtx(ctx context.Context, userID uuid.UUID) (*traineeCtx, error) {
	var tc traineeCtx
	err := s.pool.QueryRow(ctx, `
		SELECT tp.id, u.display_name, tp.student_number
		FROM trainee_profiles tp JOIN users u ON u.id = tp.user_id
		WHERE tp.user_id = $1`, userID).
		Scan(&tc.TraineeID, &tc.DisplayName, &tc.StudentNumber)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httpx.ErrForbidden
	}
	if err != nil {
		return nil, httpx.Internal(err)
	}
	err = s.pool.QueryRow(ctx, `
		SELECT timezone, low_accuracy_threshold_m, unusual_session_minutes
		FROM institution_settings LIMIT 1`).
		Scan(&tc.Timezone, &tc.LowAccuracyM, &tc.UnusualMin)
	if err != nil {
		return nil, httpx.Internal(err)
	}
	return &tc, nil
}

type siteCtx struct {
	AssignmentID uuid.UUID
	SiteID       uuid.UUID
	SiteName     string
	Latitude     float64
	Longitude    float64
	RadiusM      int
	BreakRule    BreakRule
	RequiredMin  int
}

func (s *Service) activeSite(ctx context.Context, tc *traineeCtx, today string) (*siteCtx, error) {
	var sc siteCtx
	var threshold *int
	err := s.pool.QueryRow(ctx, `
		SELECT a.id, s.id, s.name, s.latitude, s.longitude, s.allowed_radius_m,
		       a.break_rule_type, a.break_threshold_minutes, a.break_deduction_minutes,
		       a.required_minutes
		FROM ojt_assignments a
		JOIN ojt_sites s ON s.id = a.site_id
		WHERE a.trainee_id = $1
		  AND a.status = 'active'
		  AND a.start_date <= $2::date
		  AND (a.end_date IS NULL OR a.end_date >= $2::date)
		ORDER BY a.start_date DESC
		LIMIT 1`, tc.TraineeID, today).
		Scan(&sc.AssignmentID, &sc.SiteID, &sc.SiteName, &sc.Latitude, &sc.Longitude,
			&sc.RadiusM, &sc.BreakRule.Type, &threshold, &sc.BreakRule.DeductionMinutes,
			&sc.RequiredMin)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if threshold != nil {
		sc.BreakRule.ThresholdMinutes = *threshold
	}
	return &sc, nil
}

// locationFor evaluates the untrusted fix and enforces the reason rule.
func locationFor(in EvidenceInput, site *siteCtx, lowAccM float64) (LocationResult, error) {
	res := EvaluateLocation(LocationInput{
		Latitude:  in.Latitude,
		Longitude: in.Longitude,
		AccuracyM: in.GPSAccuracyM,
	}, site.Latitude, site.Longitude, site.RadiusM, lowAccM)
	if res.Status == LocationUnavailable && in.LocationExceptionReason == "" {
		return res, httpx.Unprocessable("LOCATION_REASON_REQUIRED",
			"A reason is required when location is unavailable.")
	}
	return res, nil
}

func flagsFor(status LocationStatus) []string {
	switch status {
	case LocationOutsideRadius:
		return []string{"outside_radius"}
	case LocationLowAccuracy:
		return []string{"low_accuracy"}
	case LocationUnavailable:
		return []string{"unavailable_location"}
	default:
		return []string{}
	}
}

func wmCtx(tc *traineeCtx, site *siteCtx, action string, now time.Time, loc LocationResult) WatermarkContext {
	label := "Time In"
	if action == "time_out" {
		label = "Time Out"
	}
	return WatermarkContext{
		TraineeName:    tc.DisplayName,
		StudentNumber:  tc.StudentNumber,
		Action:         label,
		ServerTime:     now,
		Timezone:       tc.Timezone,
		SiteName:       site.SiteName,
		LocationStatus: loc.Status,
	}
}

// findReplay returns the canonical attendance for an idempotent retry, or nil.
// It looks up the committed evidence row for (trainee, client_action_id).
func (s *Service) findReplay(ctx context.Context, traineeID, actionID uuid.UUID) (*AttendanceView, error) {
	var v AttendanceView
	err := s.pool.QueryRow(ctx, `
		SELECT s.id, s.attendance_date::text, s.status,
		       s.effective_time_in_at, s.effective_time_out_at, s.credited_minutes,
		       e.location_status, e.distance_from_site_m, e.gps_accuracy_m, s.flag_codes
		FROM attendance_evidence e
		JOIN attendance_sessions s ON s.id = e.attendance_id
		WHERE e.trainee_id = $1 AND e.client_action_id = $2`, traineeID, actionID).
		Scan(&v.ID, &v.Date, &v.Status, &v.OfficialTimeInAt, &v.OfficialTimeOutAt,
			&v.CreditedMinutes, &v.LocationStatus, &v.DistanceFromSiteM, &v.GPSAccuracyM, &v.Flags)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// ---- Time In ---------------------------------------------------------------

func (s *Service) TimeIn(ctx context.Context, userID uuid.UUID, in EvidenceInput) (*TimeInResult, bool, error) {
	tc, err := s.traineeCtx(ctx, userID)
	if err != nil {
		return nil, false, err
	}
	now := time.Now()
	today := MustLocalDate(now, tc.Timezone)

	site, err := s.activeSite(ctx, tc, today)
	if err != nil {
		return nil, false, httpx.Internal(err)
	}
	if site == nil {
		return nil, false, httpx.New(403, "NO_ACTIVE_ASSIGNMENT",
			"No active OJT assignment covers today.")
	}

	// Idempotent replay: committed evidence for this client action?
	if prev, err := s.findReplay(ctx, tc.TraineeID, in.ClientActionID); err != nil {
		return nil, false, httpx.Internal(err)
	} else if prev != nil {
		return &TimeInResult{Attendance: *prev}, true, nil
	}

	loc, err := locationFor(in, site, tc.LowAccuracyM)
	if err != nil {
		return nil, false, err
	}

	ev, err := StoreEvidence(ctx, s.store, in.Photo, s.maxBytes,
		wmCtx(tc, site, "time_in", now, loc))
	if err != nil {
		if ae, ok := err.(*httpx.APIError); ok {
			return nil, false, ae
		}
		return nil, false, storageUnavailable(err)
	}

	var v AttendanceView
	flags := flagsFor(loc.Status)
	txErr := s.pool.InTx(ctx, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, `
			INSERT INTO attendance_sessions
			    (trainee_id, assignment_id, attendance_date,
			     original_time_in_at, effective_time_in_at, status, flag_codes)
			VALUES ($1,$2,$3::date,$4,$4,'open',$5)
			RETURNING id`, tc.TraineeID, site.AssignmentID, today, now, flags).
			Scan(&v.ID)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO attendance_evidence
			    (attendance_id, trainee_id, action, client_action_id, server_captured_at,
			     device_captured_at, device_timezone_offset_min,
			     latitude, longitude, gps_accuracy_m, distance_from_site_m, radius_used_m,
			     location_status, location_exception_reason,
			     original_object_key, watermarked_object_key,
			     original_sha256, watermarked_sha256,
			     mime_type, size_bytes, width_px, height_px,
			     watermark_text, watermark_version)
			VALUES ($1,$2,'time_in',$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23)`,
			v.ID, tc.TraineeID, in.ClientActionID, now,
			in.DeviceCapturedAt, in.DeviceTZOffsetMin,
			in.Latitude, in.Longitude, in.GPSAccuracyM, loc.DistanceM, loc.RadiusUsedM,
			string(loc.Status), nullStr(in.LocationExceptionReason),
			ev.OriginalKey, ev.WatermarkedKey,
			ev.OriginalSHA256, ev.WatermarkSHA256,
			ev.MimeType, ev.SizeBytes, ev.Width, ev.Height,
			ev.WatermarkText, watermarkVersion)
		if err != nil {
			return err
		}
		audit.Record(ctx, tx, &userID, "attendance.time_in", "attendance_session", &v.ID, "", map[string]any{
			"location_status": string(loc.Status),
		})
		return nil
	})

	if txErr != nil {
		Cleanup(ctx, s.store, ev)
		switch {
		case db.IsUniqueViolation(txErr, "attendance_evidence_client_action_key"):
			// Same client_action_id committed concurrently — return canonical.
			if prev, err := s.findReplay(ctx, tc.TraineeID, in.ClientActionID); err == nil && prev != nil {
				return &TimeInResult{Attendance: *prev}, true, nil
			}
			return nil, false, httpx.Internal(txErr)
		case db.IsUniqueViolation(txErr, "attendance_sessions_trainee_date_key"):
			return nil, false, httpx.Conflict("ATTENDANCE_ALREADY_EXISTS_TODAY",
				"An attendance session already exists for today.")
		default:
			return nil, false, httpx.Internal(txErr)
		}
	}

	v.Date = today
	v.Status = "open"
	v.OfficialTimeInAt = now
	v.LocationStatus = string(loc.Status)
	v.DistanceFromSiteM = loc.DistanceM
	v.GPSAccuracyM = in.GPSAccuracyM
	v.Flags = flags
	return &TimeInResult{Attendance: v}, false, nil
}

// ---- Time Out ---------------------------------------------------------------

func (s *Service) TimeOut(ctx context.Context, userID, attendanceID uuid.UUID, in EvidenceInput) (*TimeOutResult, bool, error) {
	tc, err := s.traineeCtx(ctx, userID)
	if err != nil {
		return nil, false, err
	}

	// Ownership + lookup (without lock first for fast-path 404).
	var (
		traineeID    uuid.UUID
		status       string
		assignmentID uuid.UUID
		timeInAt     time.Time
	)
	err = s.pool.QueryRow(ctx, `
		SELECT trainee_id, status, assignment_id, effective_time_in_at
		FROM attendance_sessions WHERE id = $1`, attendanceID).
		Scan(&traineeID, &status, &assignmentID, &timeInAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, httpx.New(404, "ATTENDANCE_NOT_FOUND", "Attendance record not found.")
	}
	if err != nil {
		return nil, false, httpx.Internal(err)
	}
	if traineeID != tc.TraineeID {
		return nil, false, httpx.New(404, "ATTENDANCE_NOT_FOUND", "Attendance record not found.")
	}

	// Idempotent replay before doing any work.
	if prev, err := s.replayTimeOut(ctx, tc.TraineeID, in.ClientActionID, attendanceID); err != nil {
		return nil, false, httpx.Internal(err)
	} else if prev != nil {
		return prev, true, nil
	}
	if status != "open" {
		return nil, false, httpx.Conflict("ATTENDANCE_NOT_OPEN", "This attendance is already closed.")
	}

	// Site for geofence + watermark (from the assignment recorded at Time In).
	site, err := s.siteForAssignment(ctx, assignmentID)
	if err != nil {
		return nil, false, httpx.Internal(err)
	}

	now := time.Now()
	loc, err := locationFor(in, site, tc.LowAccuracyM)
	if err != nil {
		return nil, false, err
	}

	ev, err := StoreEvidence(ctx, s.store, in.Photo, s.maxBytes,
		wmCtx(tc, site, "time_out", now, loc))
	if err != nil {
		if ae, ok := err.(*httpx.APIError); ok {
			return nil, false, ae
		}
		return nil, false, storageUnavailable(err)
	}

	var (
		result    AttendanceView
		journalID *uuid.UUID
	)
	txErr := s.pool.InTx(ctx, func(tx pgx.Tx) error {
		// Lock the row — concurrent Time Outs must serialize here.
		var (
			curStatus string
			curTimeIn time.Time
			curFlags  []string
		)
		err := tx.QueryRow(ctx, `
			SELECT status, effective_time_in_at, flag_codes
			FROM attendance_sessions WHERE id = $1 FOR UPDATE`, attendanceID).
			Scan(&curStatus, &curTimeIn, &curFlags)
		if err != nil {
			return err
		}
		if curStatus != "open" {
			return errNotOpen
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO attendance_evidence
			    (attendance_id, trainee_id, action, client_action_id, server_captured_at,
			     device_captured_at, device_timezone_offset_min,
			     latitude, longitude, gps_accuracy_m, distance_from_site_m, radius_used_m,
			     location_status, location_exception_reason,
			     original_object_key, watermarked_object_key,
			     original_sha256, watermarked_sha256,
			     mime_type, size_bytes, width_px, height_px,
			     watermark_text, watermark_version)
			VALUES ($1,$2,'time_out',$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23)`,
			attendanceID, tc.TraineeID, in.ClientActionID, now,
			in.DeviceCapturedAt, in.DeviceTZOffsetMin,
			in.Latitude, in.Longitude, in.GPSAccuracyM, loc.DistanceM, loc.RadiusUsedM,
			string(loc.Status), nullStr(in.LocationExceptionReason),
			ev.OriginalKey, ev.WatermarkedKey,
			ev.OriginalSHA256, ev.WatermarkSHA256,
			ev.MimeType, ev.SizeBytes, ev.Width, ev.Height,
			ev.WatermarkText, watermarkVersion)
		if err != nil {
			return err
		}

		elapsed, _, credited := CreditedMinutes(SessionTimes{In: curTimeIn, Out: now}, site.BreakRule)
		newFlags := mergeFlags(curFlags, flagsFor(loc.Status))
		if elapsed >= tc.UnusualMin {
			newFlags = mergeFlags(newFlags, []string{"unusual_duration"})
		}
		newStatus := "valid"
		if len(newFlags) > 0 {
			newStatus = "flagged"
		}

		err = tx.QueryRow(ctx, `
			UPDATE attendance_sessions SET
			    original_time_out_at = $2, effective_time_out_at = $2,
			    credited_minutes = $3, status = $4, flag_codes = $5,
			    closed_at = $2, version = version + 1, updated_at = now()
			WHERE id = $1
			RETURNING attendance_date::text`, attendanceID, now, credited, newStatus, newFlags).
			Scan(&result.Date)
		if err != nil {
			return err
		}
		result.ID = attendanceID
		result.Status = newStatus
		result.OfficialTimeInAt = curTimeIn
		result.OfficialTimeOutAt = &now
		result.CreditedMinutes = &credited
		result.LocationStatus = string(loc.Status)
		result.DistanceFromSiteM = loc.DistanceM
		result.GPSAccuracyM = in.GPSAccuracyM
		result.Flags = newFlags

		// Journal auto-create — atomic with the attendance close so a
		// completed day can never exist without its journal.
		err = tx.QueryRow(ctx, `
			INSERT INTO daily_journals (attendance_id, trainee_id, status)
			VALUES ($1,$2,'draft')
			ON CONFLICT (attendance_id) DO NOTHING
			RETURNING id`, attendanceID, tc.TraineeID).Scan(&journalID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		if journalID == nil {
			var j uuid.UUID
			if err := tx.QueryRow(ctx,
				`SELECT id FROM daily_journals WHERE attendance_id = $1`, attendanceID).Scan(&j); err == nil {
				journalID = &j
			}
		}

		// Completion milestone — fires once when the assignment's required
		// minutes are first reached (event key dedupes).
		if err := notifications.NotifyCompletionIfReachedTx(ctx, tx,
			userID, tc.TraineeID, assignmentID); err != nil {
			return err
		}

		audit.Record(ctx, tx, &userID, "attendance.time_out", "attendance_session", &attendanceID, "", map[string]any{
			"location_status":  string(loc.Status),
			"credited_minutes": credited,
			"flags":            newFlags,
		})
		return nil
	})

	if txErr != nil {
		Cleanup(ctx, s.store, ev)
		switch {
		case errors.Is(txErr, errNotOpen):
			return nil, false, httpx.Conflict("ATTENDANCE_NOT_OPEN", "This attendance is already closed.")
		case db.IsUniqueViolation(txErr, "attendance_evidence_client_action_key"):
			if prev, err := s.replayTimeOut(ctx, tc.TraineeID, in.ClientActionID, attendanceID); err == nil && prev != nil {
				return prev, true, nil
			}
			return nil, false, httpx.Internal(txErr)
		default:
			return nil, false, httpx.Internal(txErr)
		}
	}

	prog, err := s.progress(ctx, assignmentID, site.RequiredMin)
	if err != nil {
		return nil, false, httpx.Internal(err)
	}
	out := &TimeOutResult{Attendance: result, Progress: *prog}
	if journalID != nil {
		out.Journal = &journalRef{ID: *journalID, Status: "draft"}
	}
	return out, false, nil
}

var errNotOpen = errors.New("attendance not open")

// replayTimeOut returns the canonical Time Out result when this
// client_action_id already committed a time_out evidence row.
func (s *Service) replayTimeOut(ctx context.Context, traineeID, actionID, attendanceID uuid.UUID) (*TimeOutResult, error) {
	var (
		v            AttendanceView
		timeIn       time.Time
		timeOut      *time.Time
		cred         *int
		assignmentID uuid.UUID
		requiredMin  int
	)
	err := s.pool.QueryRow(ctx, `
		SELECT s.id, s.attendance_date::text, s.status, s.effective_time_in_at,
		       s.effective_time_out_at, s.credited_minutes, s.flag_codes,
		       e.location_status, e.distance_from_site_m, e.gps_accuracy_m,
		       s.assignment_id, a.required_minutes
		FROM attendance_evidence e
		JOIN attendance_sessions s ON s.id = e.attendance_id
		JOIN ojt_assignments a ON a.id = s.assignment_id
		WHERE e.trainee_id = $1 AND e.client_action_id = $2 AND e.action = 'time_out'`,
		traineeID, actionID).
		Scan(&v.ID, &v.Date, &v.Status, &timeIn, &timeOut, &cred, &v.Flags,
			&v.LocationStatus, &v.DistanceFromSiteM, &v.GPSAccuracyM,
			&assignmentID, &requiredMin)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	// Evidence replayed against a different attendance id is not a replay of
	// this request — treat as a fresh conflict handled by the caller.
	if v.ID != attendanceID {
		return nil, nil
	}
	v.OfficialTimeInAt = timeIn
	v.OfficialTimeOutAt = timeOut
	v.CreditedMinutes = cred

	prog, err := s.progress(ctx, assignmentID, requiredMin)
	if err != nil {
		return nil, err
	}
	res := &TimeOutResult{Attendance: v, Progress: *prog}
	var jID uuid.UUID
	var jStatus string
	if err := s.pool.QueryRow(ctx,
		`SELECT id, status FROM daily_journals WHERE attendance_id = $1`, attendanceID).
		Scan(&jID, &jStatus); err == nil {
		res.Journal = &journalRef{ID: jID, Status: jStatus}
	}
	return res, nil
}

func (s *Service) siteForAssignment(ctx context.Context, assignmentID uuid.UUID) (*siteCtx, error) {
	var sc siteCtx
	var threshold *int
	err := s.pool.QueryRow(ctx, `
		SELECT a.id, s.id, s.name, s.latitude, s.longitude, s.allowed_radius_m,
		       a.break_rule_type, a.break_threshold_minutes, a.break_deduction_minutes,
		       a.required_minutes
		FROM ojt_assignments a JOIN ojt_sites s ON s.id = a.site_id
		WHERE a.id = $1`, assignmentID).
		Scan(&sc.AssignmentID, &sc.SiteID, &sc.SiteName, &sc.Latitude, &sc.Longitude,
			&sc.RadiusM, &sc.BreakRule.Type, &threshold, &sc.BreakRule.DeductionMinutes,
			&sc.RequiredMin)
	if err != nil {
		return nil, err
	}
	if threshold != nil {
		sc.BreakRule.ThresholdMinutes = *threshold
	}
	return &sc, nil
}

func (s *Service) progress(ctx context.Context, assignmentID uuid.UUID, requiredMin int) (*ProgressView, error) {
	var completed int
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(sum(credited_minutes),0)::int FROM attendance_sessions
		WHERE assignment_id = $1 AND status IN ('valid','flagged','corrected')`,
		assignmentID).Scan(&completed)
	if err != nil {
		return nil, err
	}
	return &ProgressView{
		CompletedMinutes: completed,
		RequiredMinutes:  requiredMin,
		RemainingMinutes: max(requiredMin-completed, 0),
		ProgressPercent:  ProgressPercent(completed, requiredMin),
	}, nil
}

func mergeFlags(existing, add []string) []string {
	set := map[string]bool{}
	out := []string{}
	for _, f := range append(existing, add...) {
		if !set[f] {
			set[f] = true
			out = append(out, f)
		}
	}
	return out
}

func nullStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func storageUnavailable(err error) error {
	return httpx.New(503, "STORAGE_UNAVAILABLE", "Evidence storage is unavailable. Please retry.").
		WithInternal(fmt.Errorf("storage: %w", err))
}
