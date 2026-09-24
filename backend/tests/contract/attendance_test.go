// Contract tests for /api/v1/attendance/* per contracts/api-contracts.md §4.
// These run against real Postgres + MinIO (OJT_TEST_ALLOW + OJT_TEST_S3).
package contract

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/skycode/ojt-management/backend/internal/platform/db"
	"github.com/skycode/ojt-management/backend/internal/testutil"
)

const (
	siteLat, siteLon = 14.5995, 120.9842
)

// seedTrainee creates user+profile+site+active assignment and logs the
// trainee in. Returns session cookie + CSRF token.
func seedTrainee(t *testing.T, a *fiber.App, pool *db.Pool, email string) (*http.Cookie, string, string) {
	t.Helper()
	_, traineeID := testutil.CreateTrainee(t, pool, email, "Trainee123!", "SN-"+uuid.NewString()[:8])
	siteID := testutil.CreateSite(t, pool, "Test Site", siteLat, siteLon, 200)
	testutil.CreateAssignment(t, pool, traineeID, siteID, 480*60)

	status, body, cookies := testutil.Do(t, a, "POST", "/api/v1/auth/login",
		map[string]any{"email": email, "password": "Trainee123!"}, nil, nil)
	if status != 200 {
		t.Fatalf("login failed: %d %v", status, body)
	}
	var cookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "ojt_session" {
			cookie = c
		}
	}
	csrf := body["data"].(map[string]any)["csrf_token"].(string)
	return cookie, csrf, traineeID
}

func f64(v float64) *float64 { return &v }

// doTimeIn posts a multipart Time In.
func doTimeIn(t *testing.T, a *fiber.App, cookie *http.Cookie, csrf, actionID string, lat, lon, acc *float64, reason string) (int, map[string]any) {
	t.Helper()
	fields := map[string]string{
		"client_action_id":           actionID,
		"device_captured_at":         time.Now().UTC().Format(time.RFC3339),
		"device_timezone_offset_min": "-480",
	}
	if lat != nil {
		fields["latitude"] = fmt.Sprintf("%f", *lat)
	}
	if lon != nil {
		fields["longitude"] = fmt.Sprintf("%f", *lon)
	}
	if acc != nil {
		fields["gps_accuracy_m"] = fmt.Sprintf("%f", *acc)
	}
	if reason != "" {
		fields["location_exception_reason"] = reason
	}
	return testutil.DoMultipart(t, a, "POST", "/api/v1/attendance/time-in",
		fields, "photo", "capture.jpg", testutil.TestJPEG(t),
		map[string]string{"X-CSRF-Token": csrf}, []*http.Cookie{cookie})
}

func doTimeOut(t *testing.T, a *fiber.App, cookie *http.Cookie, csrf, attendanceID, actionID string, lat, lon, acc *float64, reason string) (int, map[string]any) {
	t.Helper()
	fields := map[string]string{"client_action_id": actionID}
	if lat != nil {
		fields["latitude"] = fmt.Sprintf("%f", *lat)
	}
	if lon != nil {
		fields["longitude"] = fmt.Sprintf("%f", *lon)
	}
	if acc != nil {
		fields["gps_accuracy_m"] = fmt.Sprintf("%f", *acc)
	}
	if reason != "" {
		fields["location_exception_reason"] = reason
	}
	return testutil.DoMultipart(t, a, "POST", "/api/v1/attendance/"+attendanceID+"/time-out",
		fields, "photo", "capture.jpg", testutil.TestJPEG(t),
		map[string]string{"X-CSRF-Token": csrf}, []*http.Cookie{cookie})
}

func errCode(body map[string]any) string {
	if e, ok := body["error"].(map[string]any); ok {
		if c, ok := e["code"].(string); ok {
			return c
		}
	}
	return ""
}

func TestTimeInHappyPath(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	cookie, csrf, _ := seedTrainee(t, a, pool, "tin@x.edu")

	status, body := doTimeIn(t, a, cookie, csrf, uuid.NewString(), f64(siteLat), f64(siteLon), f64(12), "")
	if status != 201 {
		t.Fatalf("expected 201, got %d: %v", status, body)
	}
	att := body["data"].(map[string]any)["attendance"].(map[string]any)
	if att["status"] != "open" || att["official_time_in_at"] == "" {
		t.Fatalf("bad attendance: %v", att)
	}
	if att["location_status"] != "verified" {
		t.Fatalf("expected verified location, got %v", att["location_status"])
	}
	if att["id"] == "" || att["date"] == "" {
		t.Fatalf("missing id/date: %v", att)
	}
}

func TestTimeInIdempotentReplay(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	cookie, csrf, _ := seedTrainee(t, a, pool, "replay@x.edu")
	actionID := uuid.NewString()

	s1, b1 := doTimeIn(t, a, cookie, csrf, actionID, f64(siteLat), f64(siteLon), f64(10), "")
	if s1 != 201 {
		t.Fatalf("first: %d %v", s1, b1)
	}
	id1 := b1["data"].(map[string]any)["attendance"].(map[string]any)["id"]

	s2, b2 := doTimeIn(t, a, cookie, csrf, actionID, f64(siteLat), f64(siteLon), f64(10), "")
	if s2 != 200 {
		t.Fatalf("replay must be 200, got %d %v", s2, b2)
	}
	id2 := b2["data"].(map[string]any)["attendance"].(map[string]any)["id"]
	if id1 != id2 {
		t.Fatalf("replay returned different attendance: %s vs %s", id1, id2)
	}

	var n int
	if err := pool.QueryRow(context.Background(),
		`SELECT count(*) FROM attendance_sessions`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("expected exactly 1 session, got %d", n)
	}
}

func TestTimeInSecondActionSameDayConflict(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	cookie, csrf, _ := seedTrainee(t, a, pool, "dup@x.edu")

	s1, _ := doTimeIn(t, a, cookie, csrf, uuid.NewString(), f64(siteLat), f64(siteLon), f64(10), "")
	if s1 != 201 {
		t.Fatalf("first: %d", s1)
	}
	// Different client_action_id same day → 409.
	s2, b2 := doTimeIn(t, a, cookie, csrf, uuid.NewString(), f64(siteLat), f64(siteLon), f64(10), "")
	if s2 != 409 || errCode(b2) != "ATTENDANCE_ALREADY_EXISTS_TODAY" {
		t.Fatalf("expected 409 ATTENDANCE_ALREADY_EXISTS_TODAY, got %d %v", s2, b2)
	}
}

func TestTimeInConcurrentSingleSession(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	cookie, csrf, traineeID := seedTrainee(t, a, pool, "race@x.edu")

	// Two concurrent Time Ins with distinct client_action_ids must produce
	// exactly one session — the DB unique index is the backstop.
	var wg sync.WaitGroup
	statuses := make(chan int, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			s, _ := doTimeIn(t, a, cookie, csrf, uuid.NewString(),
				f64(siteLat+float64(i)*0.0001), f64(siteLon), f64(10), "")
			statuses <- s
		}(i)
	}
	wg.Wait()
	close(statuses)
	got := map[int]int{}
	for s := range statuses {
		got[s]++
	}
	if got[201] != 1 {
		t.Fatalf("expected exactly one 201, got %v", got)
	}
	var n int
	pool.QueryRow(context.Background(),
		`SELECT count(*) FROM attendance_sessions WHERE trainee_id = $1`, traineeID).Scan(&n)
	if n != 1 {
		t.Fatalf("expected 1 session, got %d", n)
	}
}

func TestTimeInNoAssignment(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	// Trainee without an assignment.
	email := "noassign@x.edu"
	testutil.CreateTrainee(t, pool, email, "Trainee123!", "SN-"+uuid.NewString()[:8])
	status, body, cookies := testutil.Do(t, a, "POST", "/api/v1/auth/login",
		map[string]any{"email": email, "password": "Trainee123!"}, nil, nil)
	if status != 200 {
		t.Fatalf("login: %d", status)
	}
	var cookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "ojt_session" {
			cookie = c
		}
	}
	csrf := body["data"].(map[string]any)["csrf_token"].(string)

	s, b := doTimeIn(t, a, cookie, csrf, uuid.NewString(), f64(siteLat), f64(siteLon), f64(10), "")
	if s != 403 || errCode(b) != "NO_ACTIVE_ASSIGNMENT" {
		t.Fatalf("expected 403 NO_ACTIVE_ASSIGNMENT, got %d %v", s, b)
	}
}

func TestTimeInUnavailableLocationRequiresReason(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	cookie, csrf, _ := seedTrainee(t, a, pool, "noreason@x.edu")

	// No lat/lon and no reason → 422 LOCATION_REASON_REQUIRED.
	s, b := doTimeIn(t, a, cookie, csrf, uuid.NewString(), nil, nil, nil, "")
	if s != 422 || errCode(b) != "LOCATION_REASON_REQUIRED" {
		t.Fatalf("expected 422 LOCATION_REASON_REQUIRED, got %d %v", s, b)
	}

	// With reason → accepted, flagged unavailable.
	s2, b2 := doTimeIn(t, a, cookie, csrf, uuid.NewString(), nil, nil, nil, "GPS permission denied")
	if s2 != 201 {
		t.Fatalf("expected 201, got %d %v", s2, b2)
	}
	att := b2["data"].(map[string]any)["attendance"].(map[string]any)
	if att["location_status"] != "unavailable" {
		t.Fatalf("expected unavailable, got %v", att["location_status"])
	}
	flags := att["flags"].([]any)
	if len(flags) == 0 || flags[0] != "unavailable_location" {
		t.Fatalf("expected unavailable_location flag, got %v", flags)
	}
}

func TestTimeInOutsideRadiusFlagged(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	cookie, csrf, _ := seedTrainee(t, a, pool, "outside@x.edu")

	// ~2.2 km away.
	s, b := doTimeIn(t, a, cookie, csrf, uuid.NewString(), f64(siteLat+0.02), f64(siteLon), f64(10), "")
	if s != 201 {
		t.Fatalf("expected 201, got %d %v", s, b)
	}
	att := b["data"].(map[string]any)["attendance"].(map[string]any)
	if att["location_status"] != "outside_radius" {
		t.Fatalf("expected outside_radius, got %v", att["location_status"])
	}
	if att["distance_from_site_m"].(float64) <= 200 {
		t.Fatalf("distance should exceed radius: %v", att["distance_from_site_m"])
	}
}

func TestTimeInInvalidImage(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	cookie, csrf, _ := seedTrainee(t, a, pool, "badimg@x.edu")

	fields := map[string]string{
		"client_action_id": uuid.NewString(),
		"latitude":         fmt.Sprintf("%f", siteLat),
		"longitude":        fmt.Sprintf("%f", siteLon),
		"gps_accuracy_m":   "10",
	}
	s, b := testutil.DoMultipart(t, a, "POST", "/api/v1/attendance/time-in",
		fields, "photo", "fake.jpg", []byte("this is not an image"),
		map[string]string{"X-CSRF-Token": csrf}, []*http.Cookie{cookie})
	if s != 415 || errCode(b) != "INVALID_IMAGE" {
		t.Fatalf("expected 415 INVALID_IMAGE, got %d %v", s, b)
	}
}

func TestTimeOutHappyPath(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	cookie, csrf, _ := seedTrainee(t, a, pool, "tout@x.edu")

	s1, b1 := doTimeIn(t, a, cookie, csrf, uuid.NewString(), f64(siteLat), f64(siteLon), f64(10), "")
	if s1 != 201 {
		t.Fatalf("time in: %d %v", s1, b1)
	}
	attID := b1["data"].(map[string]any)["attendance"].(map[string]any)["id"].(string)

	s2, b2 := doTimeOut(t, a, cookie, csrf, attID, uuid.NewString(), f64(siteLat), f64(siteLon), f64(10), "")
	if s2 != 200 {
		t.Fatalf("time out: %d %v", s2, b2)
	}
	data := b2["data"].(map[string]any)
	att := data["attendance"].(map[string]any)
	if att["status"] != "valid" || att["official_time_out_at"] == "" {
		t.Fatalf("bad close: %v", att)
	}
	if att["credited_minutes"].(float64) < 0 {
		t.Fatalf("credited must be >= 0: %v", att)
	}
	prog := data["progress"].(map[string]any)
	if prog["required_minutes"].(float64) != 480*60 {
		t.Fatalf("bad progress: %v", prog)
	}
	journal := data["journal"].(map[string]any)
	if journal["id"] == "" || journal["status"] != "draft" {
		t.Fatalf("expected draft journal, got %v", journal)
	}
}

func TestTimeOutReplay(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	cookie, csrf, _ := seedTrainee(t, a, pool, "toutreplay@x.edu")

	_, b1 := doTimeIn(t, a, cookie, csrf, uuid.NewString(), f64(siteLat), f64(siteLon), f64(10), "")
	attID := b1["data"].(map[string]any)["attendance"].(map[string]any)["id"].(string)

	actionID := uuid.NewString()
	s1, r1 := doTimeOut(t, a, cookie, csrf, attID, actionID, f64(siteLat), f64(siteLon), f64(10), "")
	if s1 != 200 {
		t.Fatalf("first timeout: %d %v", s1, r1)
	}
	s2, r2 := doTimeOut(t, a, cookie, csrf, attID, actionID, f64(siteLat), f64(siteLon), f64(10), "")
	if s2 != 200 {
		t.Fatalf("replay must be 200, got %d %v", s2, r2)
	}
	att1 := r1["data"].(map[string]any)["attendance"].(map[string]any)
	att2 := r2["data"].(map[string]any)["attendance"].(map[string]any)
	if att1["id"] != att2["id"] || att1["credited_minutes"] != att2["credited_minutes"] {
		t.Fatal("replay returned different result")
	}
}

func TestTimeOutNotOpen(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	cookie, csrf, _ := seedTrainee(t, a, pool, "notopen@x.edu")

	_, b1 := doTimeIn(t, a, cookie, csrf, uuid.NewString(), f64(siteLat), f64(siteLon), f64(10), "")
	attID := b1["data"].(map[string]any)["attendance"].(map[string]any)["id"].(string)
	doTimeOut(t, a, cookie, csrf, attID, uuid.NewString(), f64(siteLat), f64(siteLon), f64(10), "")

	// Second time out with a NEW client_action_id → 409 ATTENDANCE_NOT_OPEN.
	s, b := doTimeOut(t, a, cookie, csrf, attID, uuid.NewString(), f64(siteLat), f64(siteLon), f64(10), "")
	if s != 409 || errCode(b) != "ATTENDANCE_NOT_OPEN" {
		t.Fatalf("expected 409 ATTENDANCE_NOT_OPEN, got %d %v", s, b)
	}
}

func TestTimeOutNotFoundAndNotOwner(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	cookie, csrf, _ := seedTrainee(t, a, pool, "owner@x.edu")
	otherCookie, otherCSRF, _ := seedTrainee(t, a, pool, "other@x.edu")

	// Random ID → 404.
	s, b := doTimeOut(t, a, cookie, csrf, uuid.NewString(), uuid.NewString(), f64(siteLat), f64(siteLon), f64(10), "")
	if s != 404 || errCode(b) != "ATTENDANCE_NOT_FOUND" {
		t.Fatalf("expected 404, got %d %v", s, b)
	}

	// Other trainee's attendance → 404 (no existence leak).
	_, b1 := doTimeIn(t, a, cookie, csrf, uuid.NewString(), f64(siteLat), f64(siteLon), f64(10), "")
	attID := b1["data"].(map[string]any)["attendance"].(map[string]any)["id"].(string)
	s2, b2 := doTimeOut(t, a, otherCookie, otherCSRF, attID, uuid.NewString(), f64(siteLat), f64(siteLon), f64(10), "")
	if s2 != 404 {
		t.Fatalf("expected 404 for other-owner, got %d %v", s2, b2)
	}
}
