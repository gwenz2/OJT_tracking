package integration

import (
	"context"
	"os"
	"testing"

	"github.com/skycode/ojt-management/backend/internal/testutil"
)

// This package shares the Postgres server with tests/contract but must not
// share the database — TruncateAll would wipe contract fixtures mid-run.
func TestMain(m *testing.M) {
	if os.Getenv("OJT_TEST_DSN") == "" {
		os.Setenv("OJT_TEST_DSN",
			"host=127.0.0.1 port=5432 user=postgres password=postgres dbname=ojt_test_integration sslmode=disable")
	}
	os.Exit(m.Run())
}

// T013 — the schema builds from zero and foundation constraints hold.
func TestFoundationConstraints(t *testing.T) {
	pool := testutil.Pool(t)
	ctx := context.Background()

	expectFail := func(name, sql string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, sql, args...); err == nil {
			t.Fatalf("%s: expected constraint violation", name)
		}
	}

	// users: email unique (case-insensitive), role check
	pool.Exec(ctx, `INSERT INTO users (email, password_hash, role, display_name)
		VALUES ('a@x.edu','h','admin','A')`)
	expectFail("dup email", `INSERT INTO users (email, password_hash, role, display_name)
		VALUES ('A@X.EDU','h','admin','B')`)
	expectFail("bad role", `INSERT INTO users (email, password_hash, role, display_name)
		VALUES ('b@x.edu','h','superadmin','B')`)

	// sites: coordinate + radius bounds
	expectFail("lat>90", `INSERT INTO ojt_sites (name,address,latitude,longitude,allowed_radius_m)
		VALUES ('s','a',91,0,100)`)
	expectFail("lng>180", `INSERT INTO ojt_sites (name,address,latitude,longitude,allowed_radius_m)
		VALUES ('s','a',0,181,100)`)
	expectFail("radius 0", `INSERT INTO ojt_sites (name,address,latitude,longitude,allowed_radius_m)
		VALUES ('s','a',0,0,0)`)

	// assignments: required_minutes>0, one active per trainee, weekdays 1..7
	var uid, tid, sid string
	pool.QueryRow(ctx, `INSERT INTO users (email,password_hash,role,display_name)
		VALUES ('t@x.edu','h','trainee','T') RETURNING id`).Scan(&uid)
	pool.QueryRow(ctx, `INSERT INTO trainee_profiles (user_id, student_number)
		VALUES ($1,'S001') RETURNING id`, uid).Scan(&tid)
	pool.QueryRow(ctx, `INSERT INTO ojt_sites (name,address,latitude,longitude,allowed_radius_m)
		VALUES ('s','a',6.6,124.6,150) RETURNING id`).Scan(&sid)

	expectFail("required 0", `INSERT INTO ojt_assignments (trainee_id,site_id,start_date,required_minutes)
		VALUES ($1,$2,'2026-01-01',0)`, tid, sid)
	expectFail("bad weekday", `INSERT INTO ojt_assignments (trainee_id,site_id,start_date,required_minutes,expected_weekdays)
		VALUES ($1,$2,'2026-01-01',100,'{9}')`, tid, sid)

	pool.Exec(ctx, `INSERT INTO ojt_assignments (trainee_id,site_id,start_date,required_minutes,status)
		VALUES ($1,$2,'2026-01-01',100,'active')`, tid, sid)
	expectFail("second active assignment", `INSERT INTO ojt_assignments (trainee_id,site_id,start_date,required_minutes,status)
		VALUES ($1,$2,'2026-02-01',100,'active')`, tid, sid)

	// settings singleton + no invented retention
	var retention *int
	pool.QueryRow(ctx, `SELECT retention_days FROM institution_settings LIMIT 1`).Scan(&retention)
	if retention != nil {
		t.Fatal("retention_days must default to NULL (policy not yet approved)")
	}
	expectFail("second settings row", `INSERT INTO institution_settings (id) VALUES (gen_random_uuid())`)
}
