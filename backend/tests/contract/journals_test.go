// Contract tests for /api/v1/journals/* + /staff/journals/* per
// contracts/api-contracts.md §5 (T093–T096). Real Postgres + MinIO.
package contract

import (
	"context"
	"net/http"
	"sync"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/skycode/ojt-management/backend/internal/platform/db"
	"github.com/skycode/ojt-management/backend/internal/testutil"
)

const narrative = "Deployed and tested the new inventory endpoints today; learned how row locking prevents double submissions."

// scopeTrainee grants the coordinator visibility over the trainee.
func scopeTrainee(t *testing.T, pool *db.Pool, coordEmail, traineeID string) {
	t.Helper()
	var coordID string
	if err := pool.QueryRow(context.Background(),
		`SELECT id FROM users WHERE email = $1`, coordEmail).Scan(&coordID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(context.Background(), `
		INSERT INTO coordinator_scopes (coordinator_user_id, trainee_id)
		VALUES ($1,$2) ON CONFLICT DO NOTHING`, coordID, traineeID); err != nil {
		t.Fatal(err)
	}
}

// completedDay performs Time In + Time Out and returns (journalID, attendanceID).
func completedDay(t *testing.T, a *fiber.App, cookie *http.Cookie, csrf string) (string, string) {
	t.Helper()
	s1, b1 := doTimeIn(t, a, cookie, csrf, uuid.NewString(), f64(siteLat), f64(siteLon), f64(10), "")
	if s1 != 201 {
		t.Fatalf("time in: %d %v", s1, b1)
	}
	attID := b1["data"].(map[string]any)["attendance"].(map[string]any)["id"].(string)
	s2, b2 := doTimeOut(t, a, cookie, csrf, attID, uuid.NewString(), f64(siteLat), f64(siteLon), f64(10), "")
	if s2 != 200 {
		t.Fatalf("time out: %d %v", s2, b2)
	}
	jid := b2["data"].(map[string]any)["journal"].(map[string]any)["id"].(string)
	return jid, attID
}

func doJSON(t *testing.T, a *fiber.App, method, path string, body any, cookie *http.Cookie, csrf string) (int, map[string]any) {
	t.Helper()
	status, resp, _ := testutil.Do(t, a, method, path, body,
		map[string]string{"X-CSRF-Token": csrf}, []*http.Cookie{cookie})
	return status, resp
}

func TestJournalAutoCreatedAndOwned(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	cookie, csrf, _ := seedTrainee(t, a, pool, "j1@x.edu")

	jid, attID := completedDay(t, a, cookie, csrf)

	s, b := doJSON(t, a, "GET", "/api/v1/journals/"+jid, nil, cookie, csrf)
	if s != 200 {
		t.Fatalf("get journal: %d %v", s, b)
	}
	j := b["data"].(map[string]any)
	if j["status"] != "draft" || j["attendance_id"] != attID {
		t.Fatalf("bad journal: %v", j)
	}
	att := j["attendance"].(map[string]any)
	if att["site_name"] == "" || att["time_in_at"] == "" {
		t.Fatalf("missing attendance summary: %v", att)
	}
}

func TestJournalDraftSubmitLifecycle(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	cookie, csrf, _ := seedTrainee(t, a, pool, "j2@x.edu")
	jid, _ := completedDay(t, a, cookie, csrf)

	// Save draft.
	s, b := doJSON(t, a, "PUT", "/api/v1/journals/"+jid+"/draft",
		map[string]any{"narrative": "partial draft"}, cookie, csrf)
	if s != 200 || b["data"].(map[string]any)["status"] != "draft" {
		t.Fatalf("draft save: %d %v", s, b)
	}

	// Submit with too-short narrative → 422.
	s, b = doJSON(t, a, "POST", "/api/v1/journals/"+jid+"/submit",
		map[string]any{"narrative": "short"}, cookie, csrf)
	if s != 422 || errCode(b) != "NARRATIVE_REQUIRED" {
		t.Fatalf("expected 422 NARRATIVE_REQUIRED, got %d %v", s, b)
	}

	// Submit properly → submitted.
	s, b = doJSON(t, a, "POST", "/api/v1/journals/"+jid+"/submit",
		map[string]any{"narrative": narrative}, cookie, csrf)
	if s != 200 || b["data"].(map[string]any)["status"] != "submitted" {
		t.Fatalf("submit: %d %v", s, b)
	}
	if b["data"].(map[string]any)["submitted_at"] == nil {
		t.Fatal("submitted_at must be set")
	}

	// Draft edit after submit → 409 JOURNAL_NOT_EDITABLE.
	s, b = doJSON(t, a, "PUT", "/api/v1/journals/"+jid+"/draft",
		map[string]any{"narrative": "try again"}, cookie, csrf)
	if s != 409 || errCode(b) != "JOURNAL_NOT_EDITABLE" {
		t.Fatalf("expected 409 JOURNAL_NOT_EDITABLE, got %d %v", s, b)
	}
}

func TestJournalEvidenceUploadDelete(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	cookie, csrf, _ := seedTrainee(t, a, pool, "j3@x.edu")
	jid, _ := completedDay(t, a, cookie, csrf)

	// Upload a supporting image.
	s, b := testutil.DoMultipart(t, a, "POST", "/api/v1/journals/"+jid+"/evidence",
		nil, "photo", "note.jpg", testutil.TestJPEG(t),
		map[string]string{"X-CSRF-Token": csrf}, []*http.Cookie{cookie})
	if s != 201 {
		t.Fatalf("evidence upload: %d %v", s, b)
	}
	eid := b["data"].(map[string]any)["id"].(string)

	// Detail lists it.
	s, b = doJSON(t, a, "GET", "/api/v1/journals/"+jid, nil, cookie, csrf)
	ev := b["data"].(map[string]any)["evidence"].([]any)
	if s != 200 || len(ev) != 1 {
		t.Fatalf("expected 1 evidence item: %d %v", s, b)
	}

	// Delete it.
	s, _ = doJSON(t, a, "DELETE", "/api/v1/journals/"+jid+"/evidence/"+eid, nil, cookie, csrf)
	if s != 200 {
		t.Fatalf("delete evidence: %d", s)
	}
	s, b = doJSON(t, a, "GET", "/api/v1/journals/"+jid, nil, cookie, csrf)
	if len(b["data"].(map[string]any)["evidence"].([]any)) != 0 {
		t.Fatal("evidence should be gone")
	}

	// After submit, evidence upload is locked → 409.
	doJSON(t, a, "POST", "/api/v1/journals/"+jid+"/submit",
		map[string]any{"narrative": narrative}, cookie, csrf)
	s, b = testutil.DoMultipart(t, a, "POST", "/api/v1/journals/"+jid+"/evidence",
		nil, "photo", "note.jpg", testutil.TestJPEG(t),
		map[string]string{"X-CSRF-Token": csrf}, []*http.Cookie{cookie})
	if s != 409 || errCode(b) != "JOURNAL_NOT_EDITABLE" {
		t.Fatalf("expected 409 JOURNAL_NOT_EDITABLE, got %d %v", s, b)
	}
}

func TestJournalReviewFlow(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	cookie, csrf, traineeID := seedTrainee(t, a, pool, "j4@x.edu")
	jid, _ := completedDay(t, a, cookie, csrf)

	coordCookie, coordCSRF := testutil.LoginAs(t, a, pool, "coord-j@x.edu", "pass-12345", "coordinator", "Coord")
	scopeTrainee(t, pool, "coord-j@x.edu", traineeID)

	// Reviewing a draft → 409 JOURNAL_NOT_SUBMITTED.
	s, b := doJSON(t, a, "POST", "/api/v1/staff/journals/"+jid+"/review",
		map[string]any{"decision": "reviewed"}, coordCookie, coordCSRF)
	if s != 409 || errCode(b) != "JOURNAL_NOT_SUBMITTED" {
		t.Fatalf("expected 409 JOURNAL_NOT_SUBMITTED, got %d %v", s, b)
	}

	doJSON(t, a, "POST", "/api/v1/journals/"+jid+"/submit",
		map[string]any{"narrative": narrative}, cookie, csrf)

	// needs_revision without comment → 422.
	s, b = doJSON(t, a, "POST", "/api/v1/staff/journals/"+jid+"/review",
		map[string]any{"decision": "needs_revision"}, coordCookie, coordCSRF)
	if s != 422 || errCode(b) != "REVIEW_COMMENT_REQUIRED" {
		t.Fatalf("expected 422 REVIEW_COMMENT_REQUIRED, got %d %v", s, b)
	}

	// needs_revision with comment → state + history.
	s, b = doJSON(t, a, "POST", "/api/v1/staff/journals/"+jid+"/review",
		map[string]any{"decision": "needs_revision", "comment": "Add details on tasks."},
		coordCookie, coordCSRF)
	if s != 200 || b["data"].(map[string]any)["status"] != "needs_revision" {
		t.Fatalf("review: %d %v", s, b)
	}
	if b["data"].(map[string]any)["revision_count"].(float64) != 1 {
		t.Fatalf("revision_count should be 1: %v", b["data"])
	}

	// Trainee revises + resubmits.
	doJSON(t, a, "PUT", "/api/v1/journals/"+jid+"/draft",
		map[string]any{"narrative": narrative + " Added mentor pairing notes."}, cookie, csrf)
	doJSON(t, a, "POST", "/api/v1/journals/"+jid+"/submit",
		map[string]any{"narrative": narrative + " Added mentor pairing notes."}, cookie, csrf)

	// Approve → reviewed, history retains both decisions.
	s, b = doJSON(t, a, "POST", "/api/v1/staff/journals/"+jid+"/review",
		map[string]any{"decision": "reviewed"}, coordCookie, coordCSRF)
	if s != 200 || b["data"].(map[string]any)["status"] != "reviewed" {
		t.Fatalf("approve: %d %v", s, b)
	}
	hist := b["data"].(map[string]any)["review_history"].([]any)
	if len(hist) != 2 {
		t.Fatalf("expected 2 review entries, got %v", hist)
	}
}

func TestJournalIDORAndScope(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	cookie, csrf, traineeID := seedTrainee(t, a, pool, "j5@x.edu")
	jid, _ := completedDay(t, a, cookie, csrf)
	otherCookie, otherCSRF, _ := seedTrainee(t, a, pool, "j6@x.edu")

	// Another trainee → 404 on all access paths.
	for _, m := range []string{"GET"} {
		s, _ := doJSON(t, a, m, "/api/v1/journals/"+jid, nil, otherCookie, otherCSRF)
		if s != 404 {
			t.Fatalf("other-owner %s must 404, got %d", m, s)
		}
	}
	s, _ := doJSON(t, a, "PUT", "/api/v1/journals/"+jid+"/draft",
		map[string]any{"narrative": "x"}, otherCookie, otherCSRF)
	if s != 404 {
		t.Fatalf("other-owner draft must 404, got %d", s)
	}

	// Out-of-scope coordinator → 404; staff queue must not list it.
	unscopedCookie, unscopedCSRF := testutil.LoginAs(t, a, pool, "unscoped@x.edu", "pass-12345", "coordinator", "U")
	s, _ = doJSON(t, a, "GET", "/api/v1/journals/"+jid, nil, unscopedCookie, unscopedCSRF)
	if s != 404 {
		t.Fatalf("unscoped staff must 404, got %d", s)
	}
	s, b := doJSON(t, a, "GET", "/api/v1/staff/journals", nil, unscopedCookie, unscopedCSRF)
	if s != 200 || len(b["data"].([]any)) != 0 {
		t.Fatalf("unscoped queue must be empty: %d %v", s, b)
	}

	// In-scope coordinator sees it in the queue.
	scopedCookie, scopedCSRF := testutil.LoginAs(t, a, pool, "scoped@x.edu", "pass-12345", "coordinator", "S")
	scopeTrainee(t, pool, "scoped@x.edu", traineeID)
	s, b = doJSON(t, a, "GET", "/api/v1/staff/journals", nil, scopedCookie, scopedCSRF)
	if s != 200 || len(b["data"].([]any)) != 1 {
		t.Fatalf("scoped queue must list 1 journal: %d %v", s, b)
	}
}

func TestJournalConcurrentReviewSerializes(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	cookie, csrf, traineeID := seedTrainee(t, a, pool, "j8@x.edu")
	jid, _ := completedDay(t, a, cookie, csrf)
	doJSON(t, a, "POST", "/api/v1/journals/"+jid+"/submit",
		map[string]any{"narrative": narrative}, cookie, csrf)

	coordCookie, coordCSRF := testutil.LoginAs(t, a, pool, "race-coord@x.edu", "pass-12345", "coordinator", "RC")
	scopeTrainee(t, pool, "race-coord@x.edu", traineeID)

	// Two concurrent reviews of the same submitted journal: the row lock
	// must serialize them — exactly one wins, the loser sees 409.
	var wg sync.WaitGroup
	statuses := make(chan int, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			decision := "reviewed"
			if i == 1 {
				decision = "needs_revision"
			}
			s, _ := doJSON(t, a, "POST", "/api/v1/staff/journals/"+jid+"/review",
				map[string]any{"decision": decision, "comment": "c"}, coordCookie, coordCSRF)
			statuses <- s
		}(i)
	}
	wg.Wait()
	close(statuses)
	var ok, conflict int
	for s := range statuses {
		if s == 200 {
			ok++
		}
		if s == 409 {
			conflict++
		}
	}
	if ok != 1 || conflict != 1 {
		t.Fatalf("expected one 200 + one 409, got ok=%d conflict=%d", ok, conflict)
	}

	// Exactly one review row must exist — no double application.
	var n int
	pool.QueryRow(context.Background(),
		`SELECT count(*) FROM journal_reviews WHERE journal_id = $1`, jid).Scan(&n)
	if n != 1 {
		t.Fatalf("expected 1 review row, got %d", n)
	}
}

func TestJournalTraineeCannotReview(t *testing.T) {
	a, pool, _ := testutil.AppWithStore(t)
	cookie, csrf, _ := seedTrainee(t, a, pool, "j7@x.edu")
	jid, _ := completedDay(t, a, cookie, csrf)
	doJSON(t, a, "POST", "/api/v1/journals/"+jid+"/submit",
		map[string]any{"narrative": narrative}, cookie, csrf)

	// Trainee hitting the staff endpoint → 403 (role middleware).
	s, _ := doJSON(t, a, "POST", "/api/v1/staff/journals/"+jid+"/review",
		map[string]any{"decision": "reviewed"}, cookie, csrf)
	if s != 403 {
		t.Fatalf("trainee review must be 403, got %d", s)
	}
}
