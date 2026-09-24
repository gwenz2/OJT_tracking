// Contract tests for correction requests/decisions (api-contracts §6, T105/T108).
package contract

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/skycode/ojt-management/backend/internal/testutil"
)

// closedDay performs Time In + Time Out; returns attendanceID.
func closedDay(t *testing.T, a *fiber.App, cookie *http.Cookie, csrf string) string {
	t.Helper()
	s, b := doTimeIn(t, a, cookie, csrf, uuid.NewString(), f64(siteLat), f64(siteLon), f64(10), "")
	if s != 201 {
		t.Fatalf("time in: %d %v", s, b)
	}
	attID := b["data"].(map[string]any)["attendance"].(map[string]any)["id"].(string)
	if s, b := doTimeOut(t, a, cookie, csrf, attID, uuid.NewString(), f64(siteLat), f64(siteLon), f64(10), ""); s != 200 {
		t.Fatalf("time out: %d %v", s, b)
	}
	return attID
}

func proposedOut(hoursAgo float64) string {
	return time.Now().Add(-time.Duration(hoursAgo * float64(time.Hour))).UTC().Format(time.RFC3339)
}

func TestCorrectionRequestAndApprove(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	cookie, csrf, traineeID := seedTrainee(t, a, pool, "c1@x.edu")
	attID := closedDay(t, a, cookie, csrf)

	// Capture original timestamps — approval must never rewrite them.
	var origIn, origOut time.Time
	pool.QueryRow(context.Background(),
		`SELECT original_time_in_at, original_time_out_at FROM attendance_sessions WHERE id = $1`,
		attID).Scan(&origIn, &origOut)

	newIn := origIn.Add(-2 * time.Hour)
	s, b := doJSON(t, a, "POST", "/api/v1/attendance/"+attID+"/corrections",
		map[string]any{
			"type":                "incorrect_time",
			"reason":              "Actually arrived two hours earlier.",
			"proposed_time_in_at": newIn.Format(time.RFC3339),
		}, cookie, csrf)
	if s != 201 {
		t.Fatalf("create correction: %d %v", s, b)
	}
	corrID := b["data"].(map[string]any)["id"].(string)
	if b["data"].(map[string]any)["status"] != "pending" {
		t.Fatalf("expected pending: %v", b["data"])
	}

	// Second request on same attendance → 409.
	s, b = doJSON(t, a, "POST", "/api/v1/attendance/"+attID+"/corrections",
		map[string]any{"type": "other", "reason": "another"}, cookie, csrf)
	if s != 409 || errCode(b) != "CORRECTION_ALREADY_PENDING" {
		t.Fatalf("expected 409 CORRECTION_ALREADY_PENDING, got %d %v", s, b)
	}

	// Trainee list shows it.
	s, b = doJSON(t, a, "GET", "/api/v1/trainee/corrections", nil, cookie, csrf)
	if s != 200 || len(b["data"].([]any)) != 1 {
		t.Fatalf("trainee corrections: %d %v", s, b)
	}

	// Scoped coordinator approves.
	coord, coordCSRF := testutil.LoginAs(t, a, pool, "coord-c@x.edu", "pass-12345", "coordinator", "C")
	scopeTrainee(t, pool, "coord-c@x.edu", traineeID)

	s, b = doJSON(t, a, "GET", "/api/v1/staff/corrections?status=pending", nil, coord, coordCSRF)
	if s != 200 || len(b["data"].([]any)) != 1 {
		t.Fatalf("staff queue: %d %v", s, b)
	}

	s, b = doJSON(t, a, "POST", "/api/v1/staff/corrections/"+corrID+"/decision",
		map[string]any{"decision": "approved", "comment": "Verified."}, coord, coordCSRF)
	if s != 200 {
		t.Fatalf("approve: %d %v", s, b)
	}

	// Effective values updated; originals untouched; status corrected.
	var effIn time.Time
	var status string
	var oIn, oOut time.Time
	pool.QueryRow(context.Background(), `
		SELECT effective_time_in_at, status, original_time_in_at, original_time_out_at
		FROM attendance_sessions WHERE id = $1`, attID).Scan(&effIn, &status, &oIn, &oOut)
	if status != "corrected" {
		t.Fatalf("expected corrected status, got %s", status)
	}
	if effIn.Sub(newIn).Abs() > time.Second {
		t.Fatalf("effective in not updated: %v vs %v", effIn, newIn)
	}
	if !oIn.Equal(origIn) || !oOut.Equal(origOut) {
		t.Fatal("original timestamps must not change")
	}

	// Adjustment row + notification exist.
	var adj int
	pool.QueryRow(context.Background(),
		`SELECT count(*) FROM attendance_adjustments WHERE correction_request_id = $1`, corrID).Scan(&adj)
	if adj != 1 {
		t.Fatalf("expected 1 adjustment, got %d", adj)
	}
	var notif int
	pool.QueryRow(context.Background(), `
		SELECT count(*) FROM notifications n
		JOIN trainee_profiles tp ON tp.user_id = n.recipient_user_id
		WHERE tp.id = $1 AND n.type = 'correction_decided'`, traineeID).Scan(&notif)
	if notif != 1 {
		t.Fatalf("expected 1 notification, got %d", notif)
	}
}

func TestCorrectionValidationAndIDOR(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	cookie, csrf, _ := seedTrainee(t, a, pool, "c2@x.edu")
	otherCookie, otherCSRF, _ := seedTrainee(t, a, pool, "c3@x.edu")
	attID := closedDay(t, a, cookie, csrf)

	// missed_time_out without proposed_time_out → 422.
	s, b := doJSON(t, a, "POST", "/api/v1/attendance/"+attID+"/corrections",
		map[string]any{"type": "missed_time_out", "reason": "forgot"}, cookie, csrf)
	if s != 422 || errCode(b) != "INVALID_PROPOSED_TIME" {
		t.Fatalf("expected 422 INVALID_PROPOSED_TIME, got %d %v", s, b)
	}

	// Proposed out before effective in → 422.
	s, b = doJSON(t, a, "POST", "/api/v1/attendance/"+attID+"/corrections",
		map[string]any{
			"type":                 "incorrect_time",
			"reason":               "x",
			"proposed_time_out_at": time.Now().Add(-72 * time.Hour).UTC().Format(time.RFC3339),
		}, cookie, csrf)
	if s != 422 {
		t.Fatalf("expected 422, got %d %v", s, b)
	}

	// Other trainee's attendance → 404.
	s, _ = doJSON(t, a, "POST", "/api/v1/attendance/"+attID+"/corrections",
		map[string]any{"type": "other", "reason": "nope"}, otherCookie, otherCSRF)
	if s != 404 {
		t.Fatalf("other-owner correction must 404, got %d", s)
	}
}

func TestCorrectionRejectAndScope(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	cookie, csrf, traineeID := seedTrainee(t, a, pool, "c4@x.edu")
	attID := closedDay(t, a, cookie, csrf)

	_, b := doJSON(t, a, "POST", "/api/v1/attendance/"+attID+"/corrections",
		map[string]any{"type": "other", "reason": "please check"}, cookie, csrf)
	corrID := b["data"].(map[string]any)["id"].(string)

	// Snapshot effective values — rejection must not alter them.
	var effInBefore, effOutBefore time.Time
	pool.QueryRow(context.Background(),
		`SELECT effective_time_in_at, effective_time_out_at FROM attendance_sessions WHERE id = $1`,
		attID).Scan(&effInBefore, &effOutBefore)

	// Unscoped coordinator → 404.
	unscoped, unscopedCSRF := testutil.LoginAs(t, a, pool, "unsc@x.edu", "pass-12345", "coordinator", "U")
	s, _ := doJSON(t, a, "POST", "/api/v1/staff/corrections/"+corrID+"/decision",
		map[string]any{"decision": "approved"}, unscoped, unscopedCSRF)
	if s != 404 {
		t.Fatalf("unscoped decide must 404, got %d", s)
	}

	// Scoped coordinator rejects.
	coord, coordCSRF := testutil.LoginAs(t, a, pool, "scoped-c@x.edu", "pass-12345", "coordinator", "S")
	scopeTrainee(t, pool, "scoped-c@x.edu", traineeID)
	s, b = doJSON(t, a, "POST", "/api/v1/staff/corrections/"+corrID+"/decision",
		map[string]any{"decision": "rejected", "comment": "Evidence contradicts proposal."},
		coord, coordCSRF)
	if s != 200 {
		t.Fatalf("reject: %d %v", s, b)
	}

	var effIn, effOut time.Time
	var status string
	pool.QueryRow(context.Background(),
		`SELECT effective_time_in_at, effective_time_out_at, status FROM attendance_sessions WHERE id = $1`,
		attID).Scan(&effIn, &effOut, &status)
	if !effIn.Equal(effInBefore) || !effOut.Equal(effOutBefore) || status == "corrected" {
		t.Fatal("rejection must leave effective values unchanged")
	}

	// Deciding again → 409 CORRECTION_NOT_PENDING.
	s, b = doJSON(t, a, "POST", "/api/v1/staff/corrections/"+corrID+"/decision",
		map[string]any{"decision": "approved"}, coord, coordCSRF)
	if s != 409 || errCode(b) != "CORRECTION_NOT_PENDING" {
		t.Fatalf("expected 409 CORRECTION_NOT_PENDING, got %d %v", s, b)
	}
}

func TestNotificationEndpoints(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	cookie, csrf, traineeID := seedTrainee(t, a, pool, "c5@x.edu")
	attID := closedDay(t, a, cookie, csrf)

	// Trigger a notification via correction + approval.
	_, b := doJSON(t, a, "POST", "/api/v1/attendance/"+attID+"/corrections",
		map[string]any{"type": "other", "reason": "check"}, cookie, csrf)
	corrID := b["data"].(map[string]any)["id"].(string)
	coord, coordCSRF := testutil.LoginAs(t, a, pool, "n-coord@x.edu", "pass-12345", "coordinator", "NC")
	scopeTrainee(t, pool, "n-coord@x.edu", traineeID)
	doJSON(t, a, "POST", "/api/v1/staff/corrections/"+corrID+"/decision",
		map[string]any{"decision": "approved"}, coord, coordCSRF)

	// Trainee sees 1 unread.
	s, b := doJSON(t, a, "GET", "/api/v1/notifications?unread=true", nil, cookie, csrf)
	if s != 200 {
		t.Fatalf("notifications: %d %v", s, b)
	}
	data := b["data"].(map[string]any)
	if data["unread_count"].(float64) != 1 {
		t.Fatalf("expected 1 unread, got %v", data)
	}
	nid := data["items"].([]any)[0].(map[string]any)["id"].(string)

	// Mark one read.
	s, _ = doJSON(t, a, "POST", "/api/v1/notifications/"+nid+"/read", nil, cookie, csrf)
	if s != 200 {
		t.Fatalf("mark read: %d", s)
	}
	s, b = doJSON(t, a, "GET", "/api/v1/notifications?unread=true", nil, cookie, csrf)
	if b["data"].(map[string]any)["unread_count"].(float64) != 0 {
		t.Fatal("unread must be 0 after read")
	}

	// read-all is idempotent.
	s, _ = doJSON(t, a, "POST", "/api/v1/notifications/read-all", nil, cookie, csrf)
	if s != 200 {
		t.Fatalf("read-all: %d", s)
	}

	// Other user's notification id → 404 for the other trainee.
	otherCookie, otherCSRF, _ := seedTrainee(t, a, pool, fmt.Sprintf("c6-%s@x.edu", uuid.NewString()[:6]))
	s, _ = doJSON(t, a, "POST", "/api/v1/notifications/"+nid+"/read", nil, otherCookie, otherCSRF)
	if s != 404 {
		t.Fatalf("other-owner notification read must 404, got %d", s)
	}
}
