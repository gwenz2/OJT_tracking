// Contract tests for /staff/reports/* (T132 reconciliation, T133 scope).
package contract

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"

	"github.com/skycode/ojt-management/backend/internal/testutil"
)

func getCSV(t *testing.T, a *fiber.App, path string, cookie *http.Cookie) (int, string) {
	t.Helper()
	req, _ := http.NewRequest("GET", path, nil)
	req.AddCookie(cookie)
	res, err := a.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	return res.StatusCode, string(body)
}

func TestReportsDailyAttendanceAndCSV(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	cookie, csrf, traineeID := seedTrainee(t, a, pool, "r1@x.edu")
	closedDay(t, a, cookie, csrf)

	admin, _ := testutil.LoginAs(t, a, pool, "r-admin@x.edu", "pass-12345", "admin", "A")

	// JSON report.
	s, b := doJSON(t, a, "GET", "/api/v1/staff/reports/daily-attendance", nil, admin, "")
	if s != 200 {
		t.Fatalf("report: %d %v", s, b)
	}
	rep := b["data"].(map[string]any)
	rows := rep["rows"].([]any)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %v", rows)
	}
	row := rows[0].(map[string]any)
	if row["status"] != "valid" || row["journal_status"] != "draft" {
		t.Fatalf("bad row: %v", row)
	}

	// CSV export — header + 1 data row, BOM stripped for the check.
	code, csvText := getCSV(t, a, "/api/v1/staff/reports/daily-attendance/export.csv", admin)
	if code != 200 {
		t.Fatalf("csv: %d", code)
	}
	csvText = strings.TrimPrefix(csvText, "\xEF\xBB\xBF")
	lines := strings.Split(strings.TrimSpace(csvText), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected header+1 row, got %d lines", len(lines))
	}
	if !strings.Contains(lines[0], "Credited Minutes") {
		t.Fatalf("bad header: %s", lines[0])
	}

	// Scope: unscoped coordinator export must be empty.
	unscoped, _ := testutil.LoginAs(t, a, pool, "r-unsc@x.edu", "pass-12345", "coordinator", "U")
	code, csvText = getCSV(t, a, "/api/v1/staff/reports/daily-attendance/export.csv", unscoped)
	if code != 200 {
		t.Fatalf("unscoped csv: %d", code)
	}
	lines = strings.Split(strings.TrimSpace(strings.TrimPrefix(csvText, "\xEF\xBB\xBF")), "\n")
	if len(lines) != 1 {
		t.Fatalf("unscoped export must have header only, got %d lines", len(lines))
	}
	_ = traineeID
}

func TestReportsUnknownAndTraineeForbidden(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	cookie, csrf, _ := seedTrainee(t, a, pool, "r2@x.edu")
	admin, adminCSRF := testutil.LoginAs(t, a, pool, "r-admin2@x.edu", "pass-12345", "admin", "A")

	// Unknown report → 404.
	s, _ := doJSON(t, a, "GET", "/api/v1/staff/reports/nope", nil, admin, adminCSRF)
	if s != 404 {
		t.Fatalf("unknown report must 404, got %d", s)
	}
	// Trainee → 403.
	s, _ = doJSON(t, a, "GET", "/api/v1/staff/reports/daily-attendance", nil, cookie, csrf)
	if s != 403 {
		t.Fatalf("trainee must get 403, got %d", s)
	}
}
