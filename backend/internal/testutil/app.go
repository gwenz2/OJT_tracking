package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/skycode/ojt-management/backend/internal/app"
	"github.com/skycode/ojt-management/backend/internal/auth"
	"github.com/skycode/ojt-management/backend/internal/config"
	"github.com/skycode/ojt-management/backend/internal/platform/db"
	"github.com/skycode/ojt-management/backend/internal/platform/storage"
)

// App builds the real Fiber application against the test database. Storage is
// nil — evidence paths are covered by dedicated tests.
func App(t *testing.T) (*fiber.App, *db.Pool) {
	t.Helper()
	pool := Pool(t)
	cfg := &config.Config{
		Env:                      "development",
		AppName:                  "OJT Test",
		BodyLimitBytes:           1 << 20,
		EvidenceUploadLimitBytes: 6 << 20,
		ReadTimeout:              10 * time.Second,
		WriteTimeout:             10 * time.Second,
	}
	cfg.Auth.SessionTTL = 12 * time.Hour
	cfg.Auth.CookieName = "ojt_session"
	cfg.Auth.CookieSecure = false
	cfg.Auth.AllowedOrigins = []string{"http://localhost:5173"}
	cfg.Auth.LoginRateLimit = 5
	cfg.Auth.LoginRateWindow = time.Minute
	cfg.Auth.AttendanceRateLimit = 50
	cfg.Auth.AttendanceRateWindow = time.Minute
	cfg.Log.Format = "console"

	a := app.Build(app.Deps{Cfg: cfg, Log: slog.New(slog.DiscardHandler), Pool: pool})
	return a, pool
}

// AppWithStore builds the app wired to a real MinIO/S3 store for evidence
// paths. Skips when OJT_TEST_S3 is unset or MinIO is unreachable.
func AppWithStore(t *testing.T) (*fiber.App, *db.Pool, *storage.Store) {
	t.Helper()
	if os.Getenv("OJT_TEST_S3") == "" {
		t.Skip("set OJT_TEST_S3=1 to run storage-backed tests")
	}
	pool := Pool(t)
	cfg := testConfig()
	store, err := storage.NewS3(cfg.S3)
	if err != nil {
		t.Fatalf("s3 client: %v", err)
	}
	if err := store.EnsureBucket(context.Background()); err != nil {
		t.Skipf("minio unreachable: %v", err)
	}
	a := app.Build(app.Deps{Cfg: cfg, Log: slog.New(slog.DiscardHandler), Pool: pool, Store: store})
	return a, pool, store
}

func testConfig() *config.Config {
	cfg := &config.Config{
		Env:                      "development",
		AppName:                  "OJT Test",
		BodyLimitBytes:           1 << 20,
		EvidenceUploadLimitBytes: 6 << 20,
		ReadTimeout:              10 * time.Second,
		WriteTimeout:             10 * time.Second,
	}
	cfg.Auth.SessionTTL = 12 * time.Hour
	cfg.Auth.CookieName = "ojt_session"
	cfg.Auth.CookieSecure = false
	cfg.Auth.AllowedOrigins = []string{"http://localhost:5173"}
	cfg.Auth.LoginRateLimit = 5
	cfg.Auth.LoginRateWindow = time.Minute
	cfg.Auth.AttendanceRateLimit = 50
	cfg.Auth.AttendanceRateWindow = time.Minute
	cfg.S3.Endpoint = envOr("OJT_TEST_S3_ENDPOINT", "127.0.0.1:9000")
	cfg.S3.AccessKey = envOr("OJT_TEST_S3_KEY", "minioadmin")
	cfg.S3.SecretKey = envOr("OJT_TEST_S3_SECRET", "minioadmin")
	cfg.S3.Bucket = envOr("OJT_TEST_S3_BUCKET", "ojt-evidence-test")
	cfg.S3.PresignTTL = 5 * time.Minute
	cfg.Log.Format = "console"
	return cfg
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// ---- HTTP helpers ---------------------------------------------------------

// Do performs an in-process request against the test app and returns status,
// decoded envelope, and Set-Cookie headers.
func Do(t *testing.T, a *fiber.App, method, path string, body any, headers map[string]string, cookies []*http.Cookie) (int, map[string]any, []*http.Cookie) {
	t.Helper()
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, path, rdr)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Origin", "http://localhost:5173")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	res, err := a.Test(req, fiber.TestConfig{Timeout: 30 * time.Second})
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer res.Body.Close()
	respBody, _ := io.ReadAll(res.Body)
	var decoded map[string]any
	if len(respBody) > 0 {
		_ = json.Unmarshal(respBody, &decoded)
	}
	return res.StatusCode, decoded, res.Cookies()
}

// DoMultipart performs a multipart/form-data request: file under `photo` plus
// plain string fields.
func DoMultipart(t *testing.T, a *fiber.App, method, path string, fields map[string]string, fileField, filename string, file []byte, headers map[string]string, cookies []*http.Cookie) (int, map[string]any) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			t.Fatalf("write field: %v", err)
		}
	}
	if file != nil {
		hdr := textproto.MIMEHeader{}
		hdr.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fileField, filename))
		hdr.Set("Content-Type", "image/jpeg")
		part, err := w.CreatePart(hdr)
		if err != nil {
			t.Fatalf("create part: %v", err)
		}
		if _, err := part.Write(file); err != nil {
			t.Fatalf("write file: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	req, err := http.NewRequest(method, path, &buf)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Origin", "http://localhost:5173")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	res, err := a.Test(req, fiber.TestConfig{Timeout: 30 * time.Second})
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer res.Body.Close()
	respBody, _ := io.ReadAll(res.Body)
	var decoded map[string]any
	if len(respBody) > 0 {
		_ = json.Unmarshal(respBody, &decoded)
	}
	return res.StatusCode, decoded
}

// ---- Domain fixtures ------------------------------------------------------

// CreateTrainee inserts user+profile and returns (userID, traineeID).
func CreateTrainee(t *testing.T, pool *db.Pool, email, password, studentNo string) (string, string) {
	t.Helper()
	userID := CreateUser(t, pool, email, password, "trainee", "Trainee "+studentNo)
	var tid string
	if err := pool.QueryRow(context.Background(), `
		INSERT INTO trainee_profiles (user_id, student_number, program, year_level)
		VALUES ($1, $2, 'BSIT', '4') RETURNING id`, userID, studentNo).Scan(&tid); err != nil {
		t.Fatalf("create trainee profile: %v", err)
	}
	return userID, tid
}

// CreateSite inserts an active site at the given coordinates.
func CreateSite(t *testing.T, pool *db.Pool, name string, lat, lon float64, radius int) string {
	t.Helper()
	var id string
	if err := pool.QueryRow(context.Background(), `
		INSERT INTO ojt_sites (name, address, latitude, longitude, allowed_radius_m)
		VALUES ($1, 'Test address', $2, $3, $4) RETURNING id`, name, lat, lon, radius).Scan(&id); err != nil {
		t.Fatalf("create site: %v", err)
	}
	return id
}

// CreateAssignment inserts an active assignment covering today.
func CreateAssignment(t *testing.T, pool *db.Pool, traineeID, siteID string, requiredMin int) string {
	t.Helper()
	var id string
	if err := pool.QueryRow(context.Background(), `
		INSERT INTO ojt_assignments
		    (trainee_id, site_id, start_date, required_minutes, expected_weekdays, status)
		VALUES ($1, $2, CURRENT_DATE - 1, $3, '{1,2,3,4,5,6,7}', 'active') RETURNING id`,
		traineeID, siteID, requiredMin).Scan(&id); err != nil {
		t.Fatalf("create assignment: %v", err)
	}
	return id
}

// TestJPEG returns a small valid JPEG payload.
func TestJPEG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 320, 240))
	for y := 0; y < 240; y++ {
		for x := 0; x < 320; x++ {
			img.Set(x, y, color.RGBA{uint8(x % 255), uint8(y % 255), 100, 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatalf("encode jpeg: %v", err)
	}
	return buf.Bytes()
}

// LoginAs creates a user directly and logs in via the real endpoint,
// returning the session cookie and CSRF token.
func LoginAs(t *testing.T, a *fiber.App, pool *db.Pool, email, password, role, name string) (*http.Cookie, string) {
	t.Helper()
	CreateUser(t, pool, email, password, role, name)
	status, body, cookies := Do(t, a, "POST", "/api/v1/auth/login",
		map[string]any{"email": email, "password": password}, nil, nil)
	if status != 200 {
		t.Fatalf("login as %s failed: %d %v", email, status, body)
	}
	var session *http.Cookie
	for _, c := range cookies {
		if c.Name == "ojt_session" {
			session = c
		}
	}
	if session == nil {
		t.Fatal("no session cookie set")
	}
	csrf := body["data"].(map[string]any)["csrf_token"].(string)
	return session, csrf
}

// Ctx returns a background context for direct DB fixture calls.
func Ctx(t *testing.T) context.Context {
	t.Helper()
	return context.Background()
}

// CreateUser inserts a user row directly (test fixture helper).
func CreateUser(t *testing.T, pool *db.Pool, email, password, role, name string) string {
	t.Helper()
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	var id string
	if err := pool.QueryRow(context.Background(), `
		INSERT INTO users (email, password_hash, role, display_name)
		VALUES ($1, $2, $3, $4) RETURNING id`,
		email, hash, role, name).Scan(&id); err != nil {
		t.Fatalf("create user: %v", err)
	}
	return id
}
