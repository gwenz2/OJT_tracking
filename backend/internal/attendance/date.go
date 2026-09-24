// Pure institution-local date logic for attendance. Server timestamps are
// stored in UTC; the attendance "date" is derived in the institution's IANA
// timezone so midnight boundaries behave like the school expects.
package attendance

import (
	"fmt"
	"time"
)

// LocalDate returns the calendar date of t in the given IANA timezone as a
// "YYYY-MM-DD" string matching the attendance_sessions.attendance_date column.
func LocalDate(t time.Time, timezone string) (string, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return "", fmt.Errorf("invalid timezone %q: %w", timezone, err)
	}
	return t.In(loc).Format("2006-01-02"), nil
}

// MustLocalDate is LocalDate for callers holding an already-validated
// timezone (e.g. institution_settings.timezone checked at config/write time).
func MustLocalDate(t time.Time, timezone string) string {
	d, err := LocalDate(t, timezone)
	if err != nil {
		return t.UTC().Format("2006-01-02")
	}
	return d
}
