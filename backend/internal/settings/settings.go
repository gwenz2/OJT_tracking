// Package settings owns the single-row institution_settings table: timezone,
// thresholds, and policy values used by attendance/journal/report rules.
package settings

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/skycode/ojt-management/backend/internal/platform/db"
	"github.com/skycode/ojt-management/backend/internal/platform/httpx"
)

type Settings struct {
	ID                        uuid.UUID `json:"id"`
	Timezone                  string    `json:"timezone"`
	LowAccuracyThresholdM     int       `json:"low_accuracy_threshold_m"`
	NearCompletionPercent     float64   `json:"near_completion_percent"`
	MissingJournalCutoffHours int       `json:"missing_journal_cutoff_hours"`
	UnusualSessionMinutes     int       `json:"unusual_session_minutes"`
	RetentionDays             *int      `json:"retention_days"`
	UpdatedAt                 time.Time `json:"updated_at"`
}

// Get loads the institution settings row (created by migration 000004).
func Get(ctx context.Context, pool *db.Pool) (*Settings, error) {
	var s Settings
	err := pool.QueryRow(ctx, `
		SELECT id, timezone, low_accuracy_threshold_m, near_completion_percent,
		       missing_journal_cutoff_hours, unusual_session_minutes, retention_days, updated_at
		FROM institution_settings ORDER BY id LIMIT 1`).
		Scan(&s.ID, &s.Timezone, &s.LowAccuracyThresholdM, &s.NearCompletionPercent,
			&s.MissingJournalCutoffHours, &s.UnusualSessionMinutes, &s.RetentionDays, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Location resolves the configured IANA timezone.
func (s *Settings) Location() (*time.Location, error) {
	loc, err := time.LoadLocation(s.Timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid institution timezone %q: %w", s.Timezone, err)
	}
	return loc, nil
}

// MustLocation returns the configured location, falling back to UTC only when
// the stored value is corrupt (an admin-set invalid timezone cannot be saved
// through Update, so this should never trigger).
func (s *Settings) MustLocation() *time.Location {
	loc, err := s.Location()
	if err != nil {
		return time.UTC
	}
	return loc
}

type Patch struct {
	Timezone                  *string  `json:"timezone"`
	LowAccuracyThresholdM     *int     `json:"low_accuracy_threshold_m"`
	NearCompletionPercent     *float64 `json:"near_completion_percent"`
	MissingJournalCutoffHours *int     `json:"missing_journal_cutoff_hours"`
	UnusualSessionMinutes     *int     `json:"unusual_session_minutes"`
	RetentionDays             *int     `json:"retention_days"`
}

// Update applies validated partial changes.
func Update(ctx context.Context, pool *db.Pool, p Patch, actorID uuid.UUID) (*Settings, error) {
	cur, err := Get(ctx, pool)
	if err != nil {
		return nil, httpx.Internal(err)
	}

	fields := map[string]string{}
	if p.Timezone != nil {
		if _, err := time.LoadLocation(*p.Timezone); err != nil {
			fields["timezone"] = "Must be a valid IANA timezone (e.g. Asia/Manila)."
		} else {
			cur.Timezone = *p.Timezone
		}
	}
	if p.LowAccuracyThresholdM != nil {
		if *p.LowAccuracyThresholdM <= 0 {
			fields["low_accuracy_threshold_m"] = "Must be a positive number of meters."
		} else {
			cur.LowAccuracyThresholdM = *p.LowAccuracyThresholdM
		}
	}
	if p.NearCompletionPercent != nil {
		if *p.NearCompletionPercent <= 0 || *p.NearCompletionPercent > 100 {
			fields["near_completion_percent"] = "Must be between 0 and 100."
		} else {
			cur.NearCompletionPercent = *p.NearCompletionPercent
		}
	}
	if p.MissingJournalCutoffHours != nil {
		if *p.MissingJournalCutoffHours < 0 {
			fields["missing_journal_cutoff_hours"] = "Must be zero or greater."
		} else {
			cur.MissingJournalCutoffHours = *p.MissingJournalCutoffHours
		}
	}
	if p.UnusualSessionMinutes != nil {
		if *p.UnusualSessionMinutes <= 0 {
			fields["unusual_session_minutes"] = "Must be a positive number of minutes."
		} else {
			cur.UnusualSessionMinutes = *p.UnusualSessionMinutes
		}
	}
	if p.RetentionDays != nil {
		if *p.RetentionDays <= 0 {
			fields["retention_days"] = "Must be a positive number of days or null."
		} else {
			cur.RetentionDays = p.RetentionDays
		}
	}
	if len(fields) > 0 {
		return nil, httpx.Validation(fields)
	}

	err = pool.QueryRow(ctx, `
		UPDATE institution_settings SET
			timezone = $2, low_accuracy_threshold_m = $3, near_completion_percent = $4,
			missing_journal_cutoff_hours = $5, unusual_session_minutes = $6,
			retention_days = $7, updated_by = $8, updated_at = now()
		WHERE id = $1
		RETURNING updated_at`,
		cur.ID, cur.Timezone, cur.LowAccuracyThresholdM, cur.NearCompletionPercent,
		cur.MissingJournalCutoffHours, cur.UnusualSessionMinutes,
		cur.RetentionDays, actorID).Scan(&cur.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httpx.ErrNotFound
		}
		return nil, httpx.Internal(err)
	}
	return cur, nil
}
