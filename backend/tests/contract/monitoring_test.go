// Contract tests for staff dashboard metrics + attendance monitoring
// (T113 reconciliation, T117 auth/pagination).
package contract

import (
	"testing"

	"github.com/skycode/ojt-management/backend/internal/testutil"
)

func TestStaffDashboardMetricsReconcile(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	cookie, csrf, traineeID := seedTrainee(t, a, pool, "m1@x.edu")
	closedDay(t, a, cookie, csrf)

	coord, coordCSRF := testutil.LoginAs(t, a, pool, "m-coord@x.edu", "pass-12345", "coordinator", "MC")
	scopeTrainee(t, pool, "m-coord@x.edu", traineeID)

	s, b := doJSON(t, a, "GET", "/api/v1/staff/dashboard", nil, coord, coordCSRF)
	if s != 200 {
		t.Fatalf("dashboard: %d %v", s, b)
	}
	m := b["data"].(map[string]any)["metrics"].(map[string]any)
	if m["active_trainees"].(float64) != 1 {
		t.Fatalf("active_trainees: %v", m)
	}
	if m["present_today"].(float64) != 1 {
		t.Fatalf("present_today should be 1: %v", m)
	}
	if m["no_attendance"].(float64) != 0 {
		t.Fatalf("no_attendance should be 0 (session exists): %v", m)
	}
	// Journal is draft but not yet past the 24h cutoff → 0 missing.
	if m["missing_journals"].(float64) != 0 {
		t.Fatalf("missing_journals should be 0 before cutoff: %v", m)
	}

	// Queues are present and consistent.
	q := b["data"].(map[string]any)["queues"].(map[string]any)
	for _, k := range []string{"pending_corrections", "recent_flags", "journals_needing_attention"} {
		if _, ok := q[k]; !ok {
			t.Fatalf("missing queue %s", k)
		}
	}
}

func TestStaffDashboardScopeIsolation(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	cookie, csrf, _ := seedTrainee(t, a, pool, "m2@x.edu")
	closedDay(t, a, cookie, csrf)

	// Unscoped coordinator sees an empty dashboard.
	unscoped, unscopedCSRF := testutil.LoginAs(t, a, pool, "m-unsc@x.edu", "pass-12345", "coordinator", "U")
	s, b := doJSON(t, a, "GET", "/api/v1/staff/dashboard", nil, unscoped, unscopedCSRF)
	if s != 200 {
		t.Fatalf("dashboard: %d", s)
	}
	m := b["data"].(map[string]any)["metrics"].(map[string]any)
	if m["active_trainees"].(float64) != 0 || m["present_today"].(float64) != 0 {
		t.Fatalf("unscoped coordinator must see zero metrics: %v", m)
	}
}

func TestStaffAttendanceListFiltersAndScope(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	cookie, csrf, traineeID := seedTrainee(t, a, pool, "m3@x.edu")
	attID := closedDay(t, a, cookie, csrf)

	coord, coordCSRF := testutil.LoginAs(t, a, pool, "m-scoped@x.edu", "pass-12345", "coordinator", "S")
	scopeTrainee(t, pool, "m-scoped@x.edu", traineeID)

	// In-scope list shows the session with journal_status.
	s, b := doJSON(t, a, "GET", "/api/v1/staff/attendance", nil, coord, coordCSRF)
	if s != 200 || len(b["data"].([]any)) != 1 {
		t.Fatalf("attendance list: %d %v", s, b)
	}
	row := b["data"].([]any)[0].(map[string]any)
	if row["id"] != attID || row["status"] != "valid" || row["journal_status"] != "draft" {
		t.Fatalf("bad row: %v", row)
	}

	// Status filter: no 'flagged' rows.
	s, b = doJSON(t, a, "GET", "/api/v1/staff/attendance?status=flagged", nil, coord, coordCSRF)
	if s != 200 || len(b["data"].([]any)) != 0 {
		t.Fatalf("status filter: %d %v", s, b)
	}

	// Unscoped coordinator sees nothing.
	unscoped, unscopedCSRF := testutil.LoginAs(t, a, pool, "m-unsc2@x.edu", "pass-12345", "coordinator", "U2")
	s, b = doJSON(t, a, "GET", "/api/v1/staff/attendance", nil, unscoped, unscopedCSRF)
	if s != 200 || len(b["data"].([]any)) != 0 {
		t.Fatalf("unscoped list must be empty: %d %v", s, b)
	}

	// Trainee hitting staff endpoint → 403.
	s, _ = doJSON(t, a, "GET", "/api/v1/staff/attendance", nil, cookie, csrf)
	if s != 403 {
		t.Fatalf("trainee must get 403, got %d", s)
	}
}

func TestAttendanceDetailCorrectionHistory(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	cookie, csrf, _ := seedTrainee(t, a, pool, "m4@x.edu")
	attID := closedDay(t, a, cookie, csrf)

	// File a correction so history is non-empty.
	_, b := doJSON(t, a, "POST", "/api/v1/attendance/"+attID+"/corrections",
		map[string]any{"type": "other", "reason": "verify"}, cookie, csrf)
	if b["data"] == nil {
		t.Fatalf("correction create: %v", b)
	}

	s, b := doJSON(t, a, "GET", "/api/v1/attendance/"+attID, nil, cookie, csrf)
	if s != 200 {
		t.Fatalf("detail: %d %v", s, b)
	}
	d := b["data"].(map[string]any)
	if d["original_time_in_at"] == "" || d["effective_time_in_at"] == "" {
		t.Fatalf("original/effective must both be present: %v", d)
	}
	hist := d["correction_history"].([]any)
	if len(hist) != 1 || hist[0].(map[string]any)["status"] != "pending" {
		t.Fatalf("correction_history: %v", hist)
	}
}
