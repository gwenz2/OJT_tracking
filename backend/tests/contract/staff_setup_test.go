// Contract/integration tests for staff setup APIs (T033): trainee/site/
// assignment CRUD, CSV import, overlap rule, and coordinator scope isolation.
package contract

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/skycode/ojt-management/backend/internal/testutil"
)

// authed wraps testutil.Do with session cookie + CSRF header.
func authed(t *testing.T, a *fiber.App, session *http.Cookie, csrf, method, path string, body any) (int, map[string]any) {
	t.Helper()
	h := map[string]string{}
	if csrf != "" {
		h["X-CSRF-Token"] = csrf
	}
	status, decoded, _ := testutil.Do(t, a, method, path, body, h, []*http.Cookie{session})
	return status, decoded
}

// doMultipart sends a multipart/form-data request with one file field.
func doMultipart(t *testing.T, a *fiber.App, method, path, field, filename string, content []byte, headers map[string]string, cookies []*http.Cookie) (int, map[string]any) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="`+field+`"; filename="`+filename+`"`)
	h.Set("Content-Type", "text/csv")
	pw, _ := w.CreatePart(h)
	pw.Write(content)
	w.Close()

	req, _ := http.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Origin", "http://localhost:5173")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	res, err := a.Test(req, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatalf("multipart test: %v", err)
	}
	defer res.Body.Close()
	rb, _ := io.ReadAll(res.Body)
	var decoded map[string]any
	_ = json.Unmarshal(rb, &decoded)
	return res.StatusCode, decoded
}

func TestSiteCRUD(t *testing.T) {
	a, pool := testutil.App(t)
	sess, csrf := testutil.LoginAs(t, a, pool, "admin@x.edu", "pass-12345", "admin", "Admin")

	// create
	status, body := authed(t, a, sess, csrf, "POST", "/api/v1/staff/sites", map[string]any{
		"name": "ABC Tech", "address": "Isulan", "latitude": 6.629, "longitude": 124.605, "allowed_radius_m": 150,
	})
	if status != 201 {
		t.Fatalf("create site: %d %v", status, body)
	}
	siteID := body["data"].(map[string]any)["id"].(string)

	// invalid coordinate rejected
	status, body = authed(t, a, sess, csrf, "POST", "/api/v1/staff/sites", map[string]any{
		"name": "Bad", "address": "x", "latitude": 91, "longitude": 0, "allowed_radius_m": 100,
	})
	if status != 400 {
		t.Fatalf("invalid lat must be 400, got %d %v", status, body)
	}

	// list + detail
	status, body = authed(t, a, sess, "", "GET", "/api/v1/staff/sites", nil)
	if status != 200 || body["meta"].(map[string]any)["total"].(float64) < 1 {
		t.Fatalf("list sites: %d %v", status, body)
	}
	status, _ = authed(t, a, sess, "", "GET", "/api/v1/staff/sites/"+siteID, nil)
	if status != 200 {
		t.Fatalf("get site: %d", status)
	}

	// update radius
	status, body = authed(t, a, sess, csrf, "PATCH", "/api/v1/staff/sites/"+siteID, map[string]any{
		"allowed_radius_m": 200,
	})
	if status != 200 || body["data"].(map[string]any)["allowed_radius_m"].(float64) != 200 {
		t.Fatalf("patch site: %d %v", status, body)
	}
}

func TestTraineeCreateAndScope(t *testing.T) {
	a, pool := testutil.App(t)
	adminSess, _ := testutil.LoginAs(t, a, pool, "admin@x.edu", "pass-12345", "admin", "Admin")
	coordSess, coordCSRF := testutil.LoginAs(t, a, pool, "coord@x.edu", "pass-12345", "coordinator", "Coord")
	otherSess, _ := testutil.LoginAs(t, a, pool, "other@x.edu", "pass-12345", "coordinator", "Other")

	// coordinator creates a trainee (auto-scoped to that coordinator)
	status, body := authed(t, a, coordSess, coordCSRF, "POST", "/api/v1/staff/trainees", map[string]any{
		"email": "s1@x.edu", "display_name": "Student One", "student_number": "S-001",
	})
	if status != 201 {
		t.Fatalf("create trainee: %d %v", status, body)
	}
	data := body["data"].(map[string]any)
	if data["temporary_password"] == "" {
		t.Fatal("create must return a temporary password")
	}
	traineeID := data["trainee"].(map[string]any)["id"].(string)

	// creating coordinator sees the trainee in their list
	status, body = authed(t, a, coordSess, "", "GET", "/api/v1/staff/trainees", nil)
	if status != 200 || body["meta"].(map[string]any)["total"].(float64) != 1 {
		t.Fatalf("coordinator list: %d %v", status, body)
	}

	// a different coordinator sees nothing and gets 404 on detail
	status, body = authed(t, a, otherSess, "", "GET", "/api/v1/staff/trainees", nil)
	if status != 200 || body["meta"].(map[string]any)["total"].(float64) != 0 {
		t.Fatalf("unscoped coordinator list must be empty: %d %v", status, body)
	}
	status, _ = authed(t, a, otherSess, "", "GET", "/api/v1/staff/trainees/"+traineeID, nil)
	if status != 404 {
		t.Fatalf("out-of-scope detail must be 404, got %d", status)
	}

	// admin sees all
	status, body = authed(t, a, adminSess, "", "GET", "/api/v1/staff/trainees", nil)
	if status != 200 || body["meta"].(map[string]any)["total"].(float64) != 1 {
		t.Fatalf("admin list: %d %v", status, body)
	}
}

func TestStaffCanAssignTraineePassword(t *testing.T) {
	a, pool := testutil.App(t)
	adminSess, adminCSRF := testutil.LoginAs(t, a, pool, "admin@x.edu", "pass-12345", "admin", "Admin")

	status, body := authed(t, a, adminSess, adminCSRF, "POST", "/api/v1/staff/trainees", map[string]any{
		"email": "resetme@x.edu", "display_name": "Reset Me", "student_number": "S-RESET",
	})
	if status != 201 {
		t.Fatalf("create trainee: %d %v", status, body)
	}
	data := body["data"].(map[string]any)
	oldPassword := data["temporary_password"].(string)
	traineeID := data["trainee"].(map[string]any)["id"].(string)

	status, body = authed(t, a, adminSess, adminCSRF, "POST", "/api/v1/staff/trainees/"+traineeID+"/password", map[string]any{
		"new_password": "assigned-pass-12345",
	})
	if status != 200 {
		t.Fatalf("assign password: %d %v", status, body)
	}

	status, _, _ = testutil.Do(t, a, "POST", "/api/v1/auth/login", map[string]any{
		"email": "resetme@x.edu", "password": oldPassword,
	}, nil, nil)
	if status != 401 {
		t.Fatalf("old password must fail after reset, got %d", status)
	}
	status, _, _ = testutil.Do(t, a, "POST", "/api/v1/auth/login", map[string]any{
		"email": "resetme@x.edu", "password": "assigned-pass-12345",
	}, nil, nil)
	if status != 200 {
		t.Fatalf("new password must login, got %d", status)
	}
}

func TestAdminCanAssignCoordinatorPassword(t *testing.T) {
	a, pool := testutil.App(t)
	adminSess, adminCSRF := testutil.LoginAs(t, a, pool, "admin@x.edu", "pass-12345", "admin", "Admin")

	status, body := authed(t, a, adminSess, adminCSRF, "POST", "/api/v1/admin/coordinators", map[string]any{
		"email": "coord-reset@x.edu", "display_name": "Coord Reset",
	})
	if status != 201 {
		t.Fatalf("create coordinator: %d %v", status, body)
	}
	data := body["data"].(map[string]any)
	oldPassword := data["temporary_password"].(string)
	coordID := data["coordinator"].(map[string]any)["id"].(string)

	status, body = authed(t, a, adminSess, adminCSRF, "POST", "/api/v1/admin/coordinators/"+coordID+"/password", map[string]any{
		"new_password": "coord-assigned-12345",
	})
	if status != 200 {
		t.Fatalf("assign coordinator password: %d %v", status, body)
	}

	status, _, _ = testutil.Do(t, a, "POST", "/api/v1/auth/login", map[string]any{
		"email": "coord-reset@x.edu", "password": oldPassword,
	}, nil, nil)
	if status != 401 {
		t.Fatalf("old coordinator password must fail after assignment, got %d", status)
	}
	status, _, _ = testutil.Do(t, a, "POST", "/api/v1/auth/login", map[string]any{
		"email": "coord-reset@x.edu", "password": "coord-assigned-12345",
	}, nil, nil)
	if status != 200 {
		t.Fatalf("new coordinator password must login, got %d", status)
	}
}

func TestCSVImport(t *testing.T) {
	a, pool := testutil.App(t)
	sess, csrf := testutil.LoginAs(t, a, pool, "admin@x.edu", "pass-12345", "admin", "Admin")

	csvData := "email,student_number,display_name,program\n" +
		"ok1@x.edu,S-100,Ok One,BSIT\n" +
		"bad-email,S-101,Bad Email,BSIT\n" +
		"ok1@x.edu,S-102,Dup Email,BSIT\n"
	status, body := doMultipart(t, a, "POST", "/api/v1/staff/trainees/import/validate",
		"file", "trainees.csv", []byte(csvData), map[string]string{"X-CSRF-Token": csrf}, []*http.Cookie{sess})
	if status != 200 {
		t.Fatalf("validate: %d %v", status, body)
	}
	data := body["data"].(map[string]any)
	if data["valid_rows"].(float64) != 1 || data["invalid_rows"].(float64) != 2 {
		t.Fatalf("expected 1 valid/2 invalid, got %v", data)
	}
	token := data["import_token"].(string)

	// commit
	status, body = authed(t, a, sess, csrf, "POST", "/api/v1/staff/trainees/import/commit",
		map[string]any{"import_token": token})
	if status != 200 {
		t.Fatalf("commit: %d %v", status, body)
	}
	if body["data"].(map[string]any)["created"].(float64) != 1 {
		t.Fatalf("expected 1 created: %v", body)
	}

	// token is single-use
	status, _ = authed(t, a, sess, csrf, "POST", "/api/v1/staff/trainees/import/commit",
		map[string]any{"import_token": token})
	if status != 400 {
		t.Fatalf("reused import token must fail, got %d", status)
	}
}

func TestAssignmentOverlapRule(t *testing.T) {
	a, pool := testutil.App(t)
	sess, csrf := testutil.LoginAs(t, a, pool, "admin@x.edu", "pass-12345", "admin", "Admin")

	// trainee + site fixtures
	_, body := authed(t, a, sess, csrf, "POST", "/api/v1/staff/trainees", map[string]any{
		"email": "ov@x.edu", "display_name": "Ov", "student_number": "S-OV",
	})
	traineeID := body["data"].(map[string]any)["trainee"].(map[string]any)["id"].(string)
	_, body = authed(t, a, sess, csrf, "POST", "/api/v1/staff/sites", map[string]any{
		"name": "S1", "address": "x", "latitude": 6.6, "longitude": 124.6, "allowed_radius_m": 150,
	})
	siteID := body["data"].(map[string]any)["id"].(string)

	mk := func(status string) (int, map[string]any) {
		return authed(t, a, sess, csrf, "POST", "/api/v1/staff/assignments", map[string]any{
			"trainee_id": traineeID, "site_id": siteID, "start_date": "2026-09-01",
			"required_minutes": 30000, "status": status,
		})
	}
	if s, _ := mk("active"); s != 201 {
		t.Fatalf("first active assignment: %d", s)
	}
	// second active assignment must conflict
	if s, body := mk("active"); s != 409 || body["error"].(map[string]any)["code"] != "ACTIVE_ASSIGNMENT_EXISTS" {
		t.Fatalf("overlap must be 409 ACTIVE_ASSIGNMENT_EXISTS, got %d %v", s, body)
	}
	// a planned one is allowed
	if s, _ := mk("planned"); s != 201 {
		t.Fatalf("planned assignment alongside active should be allowed: %d", s)
	}
	_ = uuid.New() // silence unused import if assertions change
}
