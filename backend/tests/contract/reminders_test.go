// Reminder deduplication tests (T123): repeated job runs must not duplicate
// logical reminders.
package contract

import (
	"context"
	"testing"
	"time"

	"github.com/skycode/ojt-management/backend/internal/reminders"
	"github.com/skycode/ojt-management/backend/internal/testutil"
)

func TestRemindersIdempotent(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	cookie, csrf, _ := seedTrainee(t, a, pool, "rem1@x.edu")
	closedDay(t, a, cookie, csrf)

	// Age the journal's attendance close beyond the 24h cutoff so the
	// missing-journal reminder fires.
	if _, err := pool.Exec(context.Background(), `
		UPDATE attendance_sessions SET closed_at = now() - interval '48 hours'`); err != nil {
		t.Fatal(err)
	}

	o1, j1, m1, err := reminders.Run(testutil.Ctx(t), pool, time.Now(), 12)
	if err != nil {
		t.Fatal(err)
	}
	if j1 != 1 {
		t.Fatalf("expected 1 missing-journal reminder, got %d", j1)
	}

	// Second run inserts nothing — dedup via event_key.
	o2, j2, m2, err := reminders.Run(testutil.Ctx(t), pool, time.Now(), 12)
	if err != nil {
		t.Fatal(err)
	}
	if o2+j2+m2 != 0 {
		t.Fatalf("second run must be a no-op, got open=%d journal=%d attendance=%d", o2, j2, m2)
	}
	_ = o1
	_ = m1
}

func TestReminderMissingAttendance(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	// Active assignment expecting every weekday; no session today.
	_, traineeID := testutil.CreateTrainee(t, pool, "rem2@x.edu", "Trainee123!", "SN-R2")
	siteID := testutil.CreateSite(t, pool, "Rem Site", siteLat, siteLon, 200)
	testutil.CreateAssignment(t, pool, traineeID, siteID, 480*60)
	_ = a

	// Run at a time past the default hour cutoff in Manila.
	loc, _ := time.LoadLocation("Asia/Manila")
	now := time.Now().In(loc)
	runAt := time.Date(now.Year(), now.Month(), now.Day(), 18, 0, 0, 0, loc)

	_, _, m, err := reminders.Run(testutil.Ctx(t), pool, runAt, 12)
	if err != nil {
		t.Fatal(err)
	}
	if m != 1 {
		t.Fatalf("expected 1 missing-attendance reminder, got %d", m)
	}
	// Idempotent.
	_, _, m2, _ := reminders.Run(testutil.Ctx(t), pool, runAt, 12)
	if m2 != 0 {
		t.Fatalf("dedup failed: %d", m2)
	}
}
