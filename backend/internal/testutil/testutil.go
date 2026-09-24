// Package testutil provides a real-PostgreSQL test harness. Tests skip when
// OJT_TEST_DSN is unset, so unit-only environments still pass.
package testutil

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/skycode/ojt-management/backend/internal/config"
	"github.com/skycode/ojt-management/backend/internal/platform/db"
)

// Pool returns a migrated pool against the test database.
// OJT_TEST_DSN defaults to the local docker-compose database
// "ojt_test" (created on demand from the admin database).
func Pool(t *testing.T) *db.Pool {
	t.Helper()
	dsn := os.Getenv("OJT_TEST_DSN")
	if dsn == "" {
		dsn = "host=127.0.0.1 port=5432 user=postgres password=postgres dbname=ojt_test sslmode=disable"
	}
	if os.Getenv("OJT_TEST_ALLOW") == "" {
		t.Skip("set OJT_TEST_ALLOW=1 to run DB integration tests")
	}
	ensureTestDB(t, dsn)

	cfg := config.DBConfig{MaxConns: 8, MinConns: 1, MaxConnLifetime: time.Minute, MaxConnIdleTime: time.Minute}
	// Build DSN from config fields the pool understands.
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect test db: %v", err)
	}
	p := &db.Pool{Pool: pool}
	t.Cleanup(pool.Close)
	_ = cfg

	mig := db.NewMigrator(p)
	if _, err := mig.Up(context.Background()); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	TruncateAll(t, p)
	return p
}

// ensureTestDB creates the database named in dsn if missing by connecting to
// the maintenance database.
func ensureTestDB(t *testing.T, dsn string) {
	t.Helper()
	var dbname string
	for _, part := range strings.Fields(dsn) {
		if strings.HasPrefix(part, "dbname=") {
			dbname = strings.TrimPrefix(part, "dbname=")
		}
	}
	if dbname == "" || dbname == "postgres" {
		return
	}
	adminDSN := strings.Replace(dsn, "dbname="+dbname, "dbname=postgres", 1)
	conn, err := pgx.Connect(context.Background(), adminDSN)
	if err != nil {
		t.Fatalf("connect maintenance db: %v", err)
	}
	defer conn.Close(context.Background())
	var exists bool
	if err := conn.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)`, dbname).Scan(&exists); err != nil {
		t.Fatalf("check db: %v", err)
	}
	if !exists {
		if _, err := conn.Exec(context.Background(),
			fmt.Sprintf(`CREATE DATABASE %q`, dbname)); err != nil {
			t.Fatalf("create test db: %v", err)
		}
	}
}

// TruncateAll clears every domain table, preserving schema_migrations and the
// singleton settings row.
func TruncateAll(t *testing.T, pool *db.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `
		TRUNCATE TABLE
			notifications, attendance_adjustments, correction_requests,
			journal_reviews, journal_evidence, daily_journals,
			attendance_evidence, attendance_sessions,
			coordinator_scopes, ojt_assignments, ojt_sites,
			trainee_profiles, user_sessions, users, audit_events
		RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}
	// TRUNCATE users CASCADE also clears institution_settings (updated_by FK).
	// Restore the singleton row so tests see the migrated baseline.
	_, err = pool.Exec(context.Background(), `
		INSERT INTO institution_settings (id)
		SELECT gen_random_uuid() WHERE NOT EXISTS (SELECT 1 FROM institution_settings)`)
	if err != nil {
		t.Fatalf("reseed settings: %v", err)
	}
}
