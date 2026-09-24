// Migration runner: applies backend/migrations/*.up.sql in filename order,
// recording each in schema_migrations. Down files exist for development and
// documented rollback strategy; production never auto-down-migrates.
package db

import (
	"context"
	"embed"
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Migrator struct {
	pool *Pool
}

func NewMigrator(pool *Pool) *Migrator { return &Migrator{pool: pool} }

// Up applies all pending *.up.sql migrations in order.
func (m *Migrator) Up(ctx context.Context) ([]string, error) {
	if _, err := m.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    text PRIMARY KEY,
			applied_at timestamptz NOT NULL DEFAULT now()
		)`); err != nil {
		return nil, fmt.Errorf("create schema_migrations: %w", err)
	}

	applied := map[string]bool{}
	rows, err := m.pool.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		applied[v] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("read embedded migrations: %w", err)
	}
	var ups []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".up.sql") {
			ups = append(ups, e.Name())
		}
	}
	sort.Strings(ups)

	var ran []string
	for _, name := range ups {
		version := strings.TrimSuffix(name, ".up.sql")
		if applied[version] {
			continue
		}
		sql, err := migrationsFS.ReadFile("migrations/" + name)
		if err != nil {
			return ran, err
		}
		if err := m.pool.InTx(ctx, func(tx pgx.Tx) error {
			if _, err := tx.Exec(ctx, string(sql)); err != nil {
				return fmt.Errorf("migration %s: %w", name, err)
			}
			_, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, version)
			return err
		}); err != nil {
			return ran, err
		}
		ran = append(ran, version)
	}
	return ran, nil
}

// Down rolls back the most recently applied migration using its .down.sql.
// Development aid only — destructive rollbacks are never automatic.
func (m *Migrator) Down(ctx context.Context) (string, error) {
	var version string
	err := m.pool.QueryRow(ctx,
		`SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1`).Scan(&version)
	if err != nil {
		return "", fmt.Errorf("no applied migrations or missing table: %w", err)
	}
	name := "migrations/" + version + ".down.sql"
	sql, err := migrationsFS.ReadFile(name)
	if err != nil {
		return "", fmt.Errorf("down migration %s not found: %w", name, err)
	}
	if err := m.pool.InTx(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("down migration %s: %w", name, err)
		}
		_, err := tx.Exec(ctx, `DELETE FROM schema_migrations WHERE version = $1`, version)
		return err
	}); err != nil {
		return "", err
	}
	return version, nil
}
