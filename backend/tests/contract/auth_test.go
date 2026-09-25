// Contract tests for /api/v1/auth/* per contracts/api-contracts.md §2 and §17.
package contract

import (
	"context"
	"net/http"
	"testing"

	"github.com/skycode/ojt-management/backend/internal/testutil"
)

func TestLoginHappyPath(t *testing.T) {
	a, pool := testutil.App(t)
	testutil.CreateUser(t, pool, "t@x.edu", "pass-12345", "trainee", "Trainee")

	status, body, cookies := testutil.Do(t, a, "POST", "/api/v1/auth/login",
		map[string]any{"email": "t@x.edu", "password": "pass-12345"}, nil, nil)
	if status != 200 {
		t.Fatalf("expected 200, got %d: %v", status, body)
	}
	data := body["data"].(map[string]any)
	user := data["user"].(map[string]any)
	if user["role"] != "trainee" || user["display_name"] != "Trainee" {
		t.Fatalf("unexpected user payload: %v", user)
	}
	if data["csrf_token"] == "" {
		t.Fatal("login must return csrf_token")
	}
	var session *http.Cookie
	for _, c := range cookies {
		if c.Name == "ojt_session" {
			session = c
		}
	}
	if session == nil {
		t.Fatal("session cookie not set")
	}
	if !session.HttpOnly {
		t.Fatal("session cookie must be HttpOnly")
	}
	if session.SameSite != http.SameSiteLaxMode {
		t.Fatal("session cookie must be SameSite=Lax")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	a, pool := testutil.App(t)
	testutil.CreateUser(t, pool, "t@x.edu", "right-pass", "trainee", "T")

	status, body, _ := testutil.Do(t, a, "POST", "/api/v1/auth/login",
		map[string]any{"email": "t@x.edu", "password": "wrong-pass"}, nil, nil)
	if status != 401 || body["error"].(map[string]any)["code"] != "INVALID_CREDENTIALS" {
		t.Fatalf("expected 401 INVALID_CREDENTIALS, got %d %v", status, body)
	}
}

func TestLoginUnknownEmailSameError(t *testing.T) {
	a, _ := testutil.App(t)
	status, body, _ := testutil.Do(t, a, "POST", "/api/v1/auth/login",
		map[string]any{"email": "ghost@x.edu", "password": "whatever"}, nil, nil)
	if status != 401 || body["error"].(map[string]any)["code"] != "INVALID_CREDENTIALS" {
		t.Fatalf("unknown email must produce generic 401, got %d %v", status, body)
	}
}

func TestLoginInactiveAccount(t *testing.T) {
	a, pool := testutil.App(t)
	id := testutil.CreateUser(t, pool, "i@x.edu", "pass-12345", "trainee", "Inactive")
	pool.Exec(context.Background(), `UPDATE users SET account_status='inactive' WHERE id=$1`, id)

	status, body, _ := testutil.Do(t, a, "POST", "/api/v1/auth/login",
		map[string]any{"email": "i@x.edu", "password": "pass-12345"}, nil, nil)
	if status != 403 || body["error"].(map[string]any)["code"] != "ACCOUNT_INACTIVE" {
		t.Fatalf("expected 403 ACCOUNT_INACTIVE, got %d %v", status, body)
	}
}

func TestMeUnauthenticated(t *testing.T) {
	a, _ := testutil.App(t)
	status, body, _ := testutil.Do(t, a, "GET", "/api/v1/auth/me", nil, nil, nil)
	if status != 401 || body["error"].(map[string]any)["code"] != "UNAUTHENTICATED" {
		t.Fatalf("expected 401 UNAUTHENTICATED, got %d %v", status, body)
	}
}

func TestMeAndLogoutWithCSRF(t *testing.T) {
	a, pool := testutil.App(t)
	session, csrf := testutil.LoginAs(t, a, pool, "me@x.edu", "pass-12345", "trainee", "Me")

	// me works with cookie
	status, body, _ := testutil.Do(t, a, "GET", "/api/v1/auth/me", nil, nil, []*http.Cookie{session})
	if status != 200 {
		t.Fatalf("me: %d %v", status, body)
	}

	// logout without CSRF is rejected
	status, body, _ = testutil.Do(t, a, "POST", "/api/v1/auth/logout", nil, nil, []*http.Cookie{session})
	if status != 403 || body["error"].(map[string]any)["code"] != "CSRF_REJECTED" {
		t.Fatalf("logout without csrf: %d %v", status, body)
	}

	// logout with CSRF revokes the session
	status, _, _ = testutil.Do(t, a, "POST", "/api/v1/auth/logout", nil,
		map[string]string{"X-CSRF-Token": csrf}, []*http.Cookie{session})
	if status != 200 {
		t.Fatalf("logout with csrf: %d", status)
	}
	status, _, _ = testutil.Do(t, a, "GET", "/api/v1/auth/me", nil, nil, []*http.Cookie{session})
	if status != 401 {
		t.Fatalf("session must be revoked after logout, got %d", status)
	}
}

func TestChangeOwnPassword(t *testing.T) {
	a, pool := testutil.App(t)
	session, csrf := testutil.LoginAs(t, a, pool, "me@x.edu", "pass-12345", "coordinator", "Me")

	status, body, _ := testutil.Do(t, a, "POST", "/api/v1/auth/change-password", map[string]any{
		"current_password": "wrong-pass",
		"new_password":     "new-pass-12345",
	}, map[string]string{"X-CSRF-Token": csrf}, []*http.Cookie{session})
	if status != 400 {
		t.Fatalf("wrong current password must fail, got %d %v", status, body)
	}

	status, body, _ = testutil.Do(t, a, "POST", "/api/v1/auth/change-password", map[string]any{
		"current_password": "pass-12345",
		"new_password":     "new-pass-12345",
	}, map[string]string{"X-CSRF-Token": csrf}, []*http.Cookie{session})
	if status != 200 {
		t.Fatalf("change password: %d %v", status, body)
	}

	status, _, _ = testutil.Do(t, a, "POST", "/api/v1/auth/login", map[string]any{
		"email": "me@x.edu", "password": "pass-12345",
	}, nil, nil)
	if status != 401 {
		t.Fatalf("old password must fail after change, got %d", status)
	}
	status, _, _ = testutil.Do(t, a, "POST", "/api/v1/auth/login", map[string]any{
		"email": "me@x.edu", "password": "new-pass-12345",
	}, nil, nil)
	if status != 200 {
		t.Fatalf("new password must login, got %d", status)
	}
}

func TestCrossOriginUnsafeRejected(t *testing.T) {
	a, pool := testutil.App(t)
	session, csrf := testutil.LoginAs(t, a, pool, "c@x.edu", "pass-12345", "trainee", "C")

	status, body, _ := testutil.Do(t, a, "POST", "/api/v1/auth/logout", nil,
		map[string]string{"X-CSRF-Token": csrf, "Origin": "https://evil.example"},
		[]*http.Cookie{session})
	if status != 403 || body["error"].(map[string]any)["code"] != "CSRF_REJECTED" {
		t.Fatalf("foreign origin must be rejected, got %d %v", status, body)
	}
}

func TestStaffEndpointAsTraineeForbidden(t *testing.T) {
	a, pool := testutil.App(t)
	session, _ := testutil.LoginAs(t, a, pool, "tr@x.edu", "pass-12345", "trainee", "T")

	status, body, _ := testutil.Do(t, a, "GET", "/api/v1/staff/trainees", nil, nil, []*http.Cookie{session})
	if status != 403 || body["error"].(map[string]any)["code"] != "FORBIDDEN" {
		t.Fatalf("trainee on staff route: %d %v", status, body)
	}
}

func TestLoginRateLimited(t *testing.T) {
	a, pool := testutil.App(t)
	testutil.CreateUser(t, pool, "rl@x.edu", "pass-12345", "trainee", "RL")

	var lastStatus int
	for i := 0; i < 8; i++ {
		lastStatus, _, _ = testutil.Do(t, a, "POST", "/api/v1/auth/login",
			map[string]any{"email": "rl@x.edu", "password": "bad"}, nil, nil)
	}
	if lastStatus != 429 {
		t.Fatalf("expected rate limiting after burst, got %d", lastStatus)
	}
}
