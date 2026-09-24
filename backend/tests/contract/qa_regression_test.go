// Phase 5A QA regression tests:
//
//	T140 — progress fixture: completed/remaining minutes reconcile across
//	       multiple closed sessions against required_minutes.
//	T141 — server-UTC + Asia/Manila midnight boundary: the institution-local
//	       attendance date splits at Manila midnight, not UTC midnight, and
//	       the (trainee, date) unique index enforces one session per local day.
//	T142 — concurrent Time Out and concurrent correction decisions serialize
//	       on row locks: exactly one writer wins.
package contract

import (
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/skycode/ojt-management/backend/internal/attendance"
	"github.com/skycode/ojt-management/backend/internal/platform/db"
	"github.com/skycode/ojt-management/backend/internal/testutil"
)

// insertClosedSession inserts a closed attendance session directly so progress
// fixtures can be built without driving the full evidence pipeline.
// NOTE: do not call testutil.Pool here — it truncates the database.
func insertClosedSession(t *testing.T, pool *db.Pool, traineeID, assignmentID, date string, in time.Time, credited int) {
	t.Helper()
	_, err := pool.Exec(testutil.Ctx(t), `
		INSERT INTO attendance_sessions
		    (trainee_id, assignment_id, attendance_date,
		     original_time_in_at, original_time_out_at,
		     effective_time_in_at, effective_time_out_at,
		     credited_minutes, status, closed_at)
		VALUES ($1,$2,$3,$4,$5,$4,$5,$6,'valid',$5)`,
		traineeID, assignmentID, date, in, in.Add(time.Duration(credited)*time.Minute), credited)
	if err != nil {
		t.Fatalf("insert session: %v", err)
	}
}

// T140 — progress fixture: three sessions sum to completed_minutes; remaining
// and percent reconcile; over-completion caps percent at 100 (BR-008).
func TestProgressFixtureReconciles(t *testing.T) {
	a, pool := testutil.App(t)
	testutil.TruncateAll(t, pool)

	cookie, csrf, traineeID := seedTrainee(t, a, pool, "progress@example.com")
	_ = csrf

	var assignmentID string
	if err := pool.QueryRow(testutil.Ctx(t),
		`SELECT id FROM ojt_assignments WHERE trainee_id = $1`, traineeID).Scan(&assignmentID); err != nil {
		t.Fatalf("assignment: %v", err)
	}
	// seedTrainee assigns 480*60 required minutes — shrink to 90 for the fixture.
	if _, err := pool.Exec(testutil.Ctx(t),
		`UPDATE ojt_assignments SET required_minutes = 90 WHERE id = $1`, assignmentID); err != nil {
		t.Fatalf("shrink required: %v", err)
	}

	base := time.Date(2026, 3, 2, 0, 0, 0, 0, time.UTC)
	insertClosedSession(t, pool, traineeID, assignmentID, "2026-03-02", base, 30)
	insertClosedSession(t, pool, traineeID, assignmentID, "2026-03-03", base, 30)

	status, body, _ := testutil.Do(t, a, "GET", "/api/v1/trainee/dashboard",
		nil, nil, []*http.Cookie{cookie})
	if status != 200 {
		t.Fatalf("dashboard: %d %v", status, body)
	}
	progress := body["data"].(map[string]any)["assignment"].(map[string]any)
	if got := int(progress["completed_minutes"].(float64)); got != 60 {
		t.Fatalf("completed: got %d want 60", got)
	}
	if got := int(progress["remaining_minutes"].(float64)); got != 30 {
		t.Fatalf("remaining: got %d want 30", got)
	}
	if got := int(progress["progress_percent"].(float64)); got != 66 {
		t.Fatalf("percent: got %d want 66", got)
	}

	// Third session pushes past required — percent caps at 100, remaining 0.
	insertClosedSession(t, pool, traineeID, assignmentID, "2026-03-04", base, 45)
	_, body, _ = testutil.Do(t, a, "GET", "/api/v1/trainee/dashboard",
		nil, nil, []*http.Cookie{cookie})
	progress = body["data"].(map[string]any)["assignment"].(map[string]any)
	if got := int(progress["completed_minutes"].(float64)); got != 105 {
		t.Fatalf("completed over: got %d want 105", got)
	}
	if got := int(progress["remaining_minutes"].(float64)); got != 0 {
		t.Fatalf("remaining over: got %d want 0", got)
	}
	if got := int(progress["progress_percent"].(float64)); got != 100 {
		t.Fatalf("percent over: got %d want 100 (capped)", got)
	}
}

// T141 — Manila midnight boundary: two UTC instants one minute apart on
// either side of Manila midnight produce different attendance dates and may
// coexist; a second session on the same local date is rejected by the unique
// index.
func TestAttendanceDateManilaMidnightBoundary(t *testing.T) {
	a, pool := testutil.App(t)
	testutil.TruncateAll(t, pool)

	_, _, traineeID := seedTrainee(t, a, pool, "midnight@example.com")
	var assignmentID string
	if err := pool.QueryRow(testutil.Ctx(t),
		`SELECT id FROM ojt_assignments WHERE trainee_id = $1`, traineeID).Scan(&assignmentID); err != nil {
		t.Fatalf("assignment: %v", err)
	}

	// 15:59 UTC = 23:59 Manila (same day); 16:00 UTC = 00:00 Manila (next day).
	before := time.Date(2026, 3, 1, 15, 59, 0, 0, time.UTC)
	after := time.Date(2026, 3, 1, 16, 0, 0, 0, time.UTC)
	dBefore, err := attendance.LocalDate(before, "Asia/Manila")
	if err != nil {
		t.Fatalf("LocalDate: %v", err)
	}
	dAfter, err := attendance.LocalDate(after, "Asia/Manila")
	if err != nil {
		t.Fatalf("LocalDate: %v", err)
	}
	if dBefore != "2026-03-01" || dAfter != "2026-03-02" {
		t.Fatalf("boundary dates: %s / %s", dBefore, dAfter)
	}

	insert := func(date string, inTime time.Time) error {
		_, err := pool.Exec(testutil.Ctx(t), `
			INSERT INTO attendance_sessions
			    (trainee_id, assignment_id, attendance_date,
			     original_time_in_at, effective_time_in_at, status)
			VALUES ($1,$2,$3,$4,$4,'open')`, traineeID, assignmentID, date, inTime)
		return err
	}
	if err := insert(dBefore, before); err != nil {
		t.Fatalf("insert day 1: %v", err)
	}
	if err := insert(dAfter, after); err != nil {
		t.Fatalf("sessions on adjacent Manila dates must coexist: %v", err)
	}
	// Same trainee + same local date must violate the unique index.
	if err := insert(dBefore, before.Add(2*time.Hour)); err == nil {
		t.Fatal("duplicate (trainee, attendance_date) must be rejected")
	}
}

// T142 — concurrent Time Out: exactly one request wins; the rest get a
// conflict or an idempotent replay. The row lock must serialize writers.
func TestTimeOutConcurrentSerializes(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	testutil.TruncateAll(t, pool)

	cookie, csrf, _ := seedTrainee(t, a, pool, "timeout-race@example.com")
	status, body := doTimeIn(t, a, cookie, csrf, uuid.NewString(),
		f64(siteLat), f64(siteLon), f64(10), "")
	if status != 200 && status != 201 {
		t.Fatalf("time in: %d %v", status, body)
	}
	attendanceID := body["data"].(map[string]any)["attendance"].(map[string]any)["id"].(string)

	const n = 4
	var wg sync.WaitGroup
	codes := make([]int, n)
	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// Distinct action ids → no replay shortcut; the row lock decides.
			s, _ := doTimeOut(t, a, cookie, csrf, attendanceID, uuid.NewString(),
				f64(siteLat), f64(siteLon), f64(10), "")
			codes[i] = s
		}(i)
	}
	wg.Wait()

	wins := 0
	for _, c := range codes {
		if c == 200 || c == 201 {
			wins++
		} else if c != 409 {
			t.Fatalf("unexpected status %d", c)
		}
	}
	if wins != 1 {
		t.Fatalf("expected exactly 1 successful time-out, got %d (codes %v)", wins, codes)
	}
}

// T142 — concurrent correction decisions: two staff approve the same pending
// request; exactly one decision lands, the loser gets CORRECTION_NOT_PENDING.
func TestCorrectionDecisionConcurrentSerializes(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	testutil.TruncateAll(t, pool)

	cookie, csrf, _ := seedTrainee(t, a, pool, "corr-race@example.com")
	attendanceID := closedDay(t, a, cookie, csrf)

	// Trainee files a correction.
	s, b := doJSON(t, a, "POST", "/api/v1/attendance/"+attendanceID+"/corrections",
		map[string]any{
			"type":                "incorrect_time",
			"reason":              "Clocked in late by a minute",
			"proposed_time_in_at": time.Now().Add(-3 * time.Hour).UTC().Format(time.RFC3339),
		}, cookie, csrf)
	if s != 200 && s != 201 {
		t.Fatalf("correction request: %d %v", s, b)
	}
	correctionID := b["data"].(map[string]any)["id"].(string)

	// Two admins decide concurrently.
	staffCookie, staffCSRF := testutil.LoginAs(t, a, pool, "admin1@example.com", "Admin123!", "admin", "Admin One")
	staffCookie2, staffCSRF2 := testutil.LoginAs(t, a, pool, "admin2@example.com", "Admin123!", "admin", "Admin Two")

	var wg sync.WaitGroup
	res := make([]int, 2)
	decide := func(i int, c *http.Cookie, tok string) {
		defer wg.Done()
		s2, _ := doJSON(t, a, "POST",
			"/api/v1/staff/corrections/"+correctionID+"/decision",
			map[string]any{"decision": "approved"}, c, tok)
		res[i] = s2
	}
	wg.Add(2)
	go decide(0, staffCookie, staffCSRF)
	go decide(1, staffCookie2, staffCSRF2)
	wg.Wait()

	if res[0] == res[1] {
		t.Fatalf("both decisions returned %d — lock did not serialize", res[0])
	}
	for _, s := range res {
		if s != 200 && s != 409 {
			t.Fatalf("unexpected status %d", s)
		}
	}
}
