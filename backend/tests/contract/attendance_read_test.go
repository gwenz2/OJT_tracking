// Read-path tests: today, history, detail, evidence media + IDOR/scope.
package contract

import (
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/skycode/ojt-management/backend/internal/platform/db"
	"github.com/skycode/ojt-management/backend/internal/testutil"
)

func loginRole(t *testing.T, a *fiber.App, pool *db.Pool, email, role string) (*http.Cookie, string) {
	t.Helper()
	testutil.CreateUser(t, pool, email, "Staff123!", role, "Staff")
	status, body, cookies := testutil.Do(t, a, "POST", "/api/v1/auth/login",
		map[string]any{"email": email, "password": "Staff123!"}, nil, nil)
	if status != 200 {
		t.Fatalf("login %s: %d %v", email, status, body)
	}
	var cookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "ojt_session" {
			cookie = c
		}
	}
	return cookie, body["data"].(map[string]any)["csrf_token"].(string)
}

func TestAttendanceTodayEmpty(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	cookie, _, _ := seedTrainee(t, a, pool, "today0@x.edu")

	status, body, _ := testutil.Do(t, a, "GET", "/api/v1/attendance/today", nil, nil,
		[]*http.Cookie{cookie})
	if status != 200 {
		t.Fatalf("expected 200, got %d %v", status, body)
	}
	if body["data"] != nil {
		t.Fatalf("expected data:null before time-in, got %v", body["data"])
	}
}

func TestAttendanceTodayAfterTimeIn(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	cookie, csrf, _ := seedTrainee(t, a, pool, "today1@x.edu")

	_, b := doTimeIn(t, a, cookie, csrf, uuid.NewString(), f64(siteLat), f64(siteLon), f64(10), "")
	attID := b["data"].(map[string]any)["attendance"].(map[string]any)["id"]

	status, body, _ := testutil.Do(t, a, "GET", "/api/v1/attendance/today", nil, nil,
		[]*http.Cookie{cookie})
	if status != 200 {
		t.Fatalf("today: %d %v", status, body)
	}
	data := body["data"].(map[string]any)
	if data["id"] != attID || data["status"] != "open" {
		t.Fatalf("today returned wrong session: %v", data)
	}
}

func TestAttendanceHistoryOnlyOwn(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	cookie, csrf, _ := seedTrainee(t, a, pool, "hist1@x.edu")
	otherCookie, _, _ := seedTrainee(t, a, pool, "hist2@x.edu")

	doTimeIn(t, a, cookie, csrf, uuid.NewString(), f64(siteLat), f64(siteLon), f64(10), "")

	status, body, _ := testutil.Do(t, a, "GET", "/api/v1/attendance/history", nil, nil,
		[]*http.Cookie{cookie})
	if status != 200 {
		t.Fatalf("history: %d %v", status, body)
	}
	items := body["data"].([]any)
	if len(items) != 1 {
		t.Fatalf("expected 1 history row, got %v", items)
	}
	if body["meta"].(map[string]any)["total"].(float64) != 1 {
		t.Fatalf("bad meta: %v", body["meta"])
	}

	// Other trainee sees zero rows.
	_, body2, _ := testutil.Do(t, a, "GET", "/api/v1/attendance/history", nil, nil,
		[]*http.Cookie{otherCookie})
	if len(body2["data"].([]any)) != 0 {
		t.Fatalf("other trainee must see empty history: %v", body2)
	}
}

func TestAttendanceDetailOwnerStaffScope(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	cookie, csrf, traineeID := seedTrainee(t, a, pool, "own@x.edu")
	otherCookie, _, _ := seedTrainee(t, a, pool, "stranger@x.edu")

	_, b := doTimeIn(t, a, cookie, csrf, uuid.NewString(), f64(siteLat), f64(siteLon), f64(10), "")
	attID := b["data"].(map[string]any)["attendance"].(map[string]any)["id"].(string)

	// Owner sees detail.
	s, body, _ := testutil.Do(t, a, "GET", "/api/v1/attendance/"+attID, nil, nil, []*http.Cookie{cookie})
	if s != 200 || body["data"].(map[string]any)["id"] != attID {
		t.Fatalf("owner detail: %d %v", s, body)
	}
	ev := body["data"].(map[string]any)["evidence"].([]any)
	if len(ev) != 1 {
		t.Fatalf("expected 1 evidence row, got %v", ev)
	}

	// Other trainee → 404 (no existence leak).
	s2, _, _ := testutil.Do(t, a, "GET", "/api/v1/attendance/"+attID, nil, nil, []*http.Cookie{otherCookie})
	if s2 != 404 {
		t.Fatalf("other trainee must get 404, got %d", s2)
	}

	// Coordinator WITHOUT scope → 404.
	coordCookie, _ := loginRole(t, a, pool, "coord-none@x.edu", "coordinator")
	s3, _, _ := testutil.Do(t, a, "GET", "/api/v1/attendance/"+attID, nil, nil, []*http.Cookie{coordCookie})
	if s3 != 404 {
		t.Fatalf("unscoped coordinator must get 404, got %d", s3)
	}

	// Coordinator WITH scope → 200.
	coordCookie2, _ := loginRole(t, a, pool, "coord-yes@x.edu", "coordinator")
	var coordUserID string
	pool.QueryRow(testutil.Ctx(t), `SELECT id FROM users WHERE email='coord-yes@x.edu'`).Scan(&coordUserID)
	if _, err := pool.Exec(testutil.Ctx(t),
		`INSERT INTO coordinator_scopes (coordinator_user_id, trainee_id) VALUES ($1,$2)`,
		coordUserID, traineeID); err != nil {
		t.Fatal(err)
	}
	s4, _, _ := testutil.Do(t, a, "GET", "/api/v1/attendance/"+attID, nil, nil, []*http.Cookie{coordCookie2})
	if s4 != 200 {
		t.Fatalf("scoped coordinator must get 200, got %d", s4)
	}
}

func TestEvidenceImageAuth(t *testing.T) {
	a, pool, store := testutil.AppWithStore(t)
	cookie, csrf, _ := seedTrainee(t, a, pool, "ev@x.edu")

	_, b := doTimeIn(t, a, cookie, csrf, uuid.NewString(), f64(siteLat), f64(siteLon), f64(10), "")
	attID := b["data"].(map[string]any)["attendance"].(map[string]any)["id"].(string)

	// Owner gets a redirect to a presigned URL.
	req, _ := http.NewRequest("GET", "/api/v1/attendance/"+attID+"/evidence/time_in/image?variant=watermarked", nil)
	req.AddCookie(cookie)
	res, err := a.Test(req, fiber.TestConfig{Timeout: 30 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 302 && res.StatusCode != 303 {
		t.Fatalf("expected redirect to presigned URL, got %d", res.StatusCode)
	}
	loc := res.Header.Get("Location")
	if loc == "" {
		t.Fatal("no Location header")
	}

	// The presigned URL actually fetches the object from MinIO.
	get, err := http.Get(loc)
	if err != nil {
		t.Fatalf("fetch presigned: %v", err)
	}
	defer get.Body.Close()
	if get.StatusCode != 200 || get.Header.Get("Content-Type") != "image/jpeg" {
		t.Fatalf("presigned object fetch failed: %d %s", get.StatusCode, get.Header.Get("Content-Type"))
	}

	// Anonymous → 401.
	req2, _ := http.NewRequest("GET", "/api/v1/attendance/"+attID+"/evidence/time_in/image", nil)
	res2, _ := a.Test(req2)
	if res2.StatusCode != 401 {
		t.Fatalf("anonymous must get 401, got %d", res2.StatusCode)
	}

	// Variant=original also resolves for the owner.
	req3, _ := http.NewRequest("GET", "/api/v1/attendance/"+attID+"/evidence/time_in/image?variant=original", nil)
	req3.AddCookie(cookie)
	res3, _ := a.Test(req3)
	if res3.StatusCode != 302 && res3.StatusCode != 303 {
		t.Fatalf("original variant: %d", res3.StatusCode)
	}
	_ = store
}
