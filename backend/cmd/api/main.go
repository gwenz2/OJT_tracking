// OJT Attendance & Daily Journal API.
//
// Usage:
//
//	api                 serve HTTP
//	api --migrate-up    apply pending migrations and exit
//	api --migrate-down  roll back the latest migration (dev only) and exit
//	api --seed-admin    create the initial admin account from
//	                    ADMIN_EMAIL / ADMIN_PASSWORD / ADMIN_NAME env vars
//	api --reminders     run one notification-reminder pass and exit
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/skycode/ojt-management/backend/internal/app"
	"github.com/skycode/ojt-management/backend/internal/auth"
	"github.com/skycode/ojt-management/backend/internal/config"
	"github.com/skycode/ojt-management/backend/internal/platform/db"
	"github.com/skycode/ojt-management/backend/internal/platform/observability"
	"github.com/skycode/ojt-management/backend/internal/platform/storage"
	reminderssvc "github.com/skycode/ojt-management/backend/internal/reminders"
)

func main() {
	migrateUp := flag.Bool("migrate-up", false, "apply pending database migrations and exit")
	migrateDown := flag.Bool("migrate-down", false, "roll back the latest migration (dev only) and exit")
	seedAdmin := flag.Bool("seed-admin", false, "create initial admin account and exit")
	reminders := flag.Bool("reminders", false, "run one notification reminder pass and exit")
	healthCheck := flag.Bool("health-check", false, "ping database and object storage, then exit (container healthcheck)")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "configuration error: %v\n", err)
		os.Exit(1)
	}
	log := observability.NewLogger(cfg.Log)

	ctx := context.Background()

	pool, err := db.NewPool(ctx, cfg.DB)
	if err != nil {
		log.Error("database connect failed", "error", err.Error())
		os.Exit(1)
	}
	defer pool.Shutdown()

	store, err := storage.NewS3(cfg.S3)
	if err != nil {
		log.Error("storage init failed", "error", err.Error())
		os.Exit(1)
	}

	switch {
	case *migrateUp:
		ran, err := db.NewMigrator(pool).Up(ctx)
		if err != nil {
			log.Error("migration failed", "error", err.Error())
			os.Exit(1)
		}
		fmt.Printf("applied migrations: %v\n", ran)
		return
	case *migrateDown:
		v, err := db.NewMigrator(pool).Down(ctx)
		if err != nil {
			log.Error("migration rollback failed", "error", err.Error())
			os.Exit(1)
		}
		fmt.Printf("rolled back: %s\n", v)
		return
	case *seedAdmin:
		if err := seedAdminAccount(ctx, pool, log); err != nil {
			log.Error("seed admin failed", "error", err.Error())
			os.Exit(1)
		}
		return
	case *reminders:
		open, missingJ, missingA, err := reminderssvc.Run(ctx, pool, time.Now(), 12)
		if err != nil {
			log.Error("reminders failed", "error", err.Error())
			os.Exit(1)
		}
		log.Info("reminders complete",
			"open_time_in", open, "missing_journal", missingJ, "missing_attendance", missingA)
		return
	case *healthCheck:
		if err := pool.Health(ctx); err != nil {
			log.Error("health-check: database", "error", err.Error())
			os.Exit(1)
		}
		if err := store.Health(ctx); err != nil {
			log.Error("health-check: object storage", "error", err.Error())
			os.Exit(1)
		}
		return
	}

	if !cfg.IsProduction() {
		if err := store.EnsureBucket(ctx); err != nil {
			log.Warn("object storage bucket check failed", "error", err.Error())
		}
	}

	fiberApp := app.Build(app.Deps{Cfg: cfg, Log: log, Pool: pool, Store: store})

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		shCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = fiberApp.ShutdownWithContext(shCtx)
	}()

	if err := fiberApp.Listen(fmt.Sprintf(":%d", cfg.Port)); err != nil {
		log.Error("server failed", "error", err.Error())
		os.Exit(1)
	}
}

// seedAdminAccount creates the initial administrator. Credentials come from
// the environment so they never appear in source control or CLI history.
func seedAdminAccount(ctx context.Context, pool *db.Pool, log *slog.Logger) error {
	email := os.Getenv("ADMIN_EMAIL")
	password := os.Getenv("ADMIN_PASSWORD")
	name := os.Getenv("ADMIN_NAME")
	if email == "" || password == "" {
		return fmt.Errorf("ADMIN_EMAIL and ADMIN_PASSWORD env vars are required")
	}
	if name == "" {
		name = "Administrator"
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	var id string
	err = pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, role, display_name)
		VALUES ($1, $2, 'admin', $3)
		ON CONFLICT (lower(email)) DO NOTHING
		RETURNING id`, email, hash, name).Scan(&id)
	if err != nil {
		if err.Error() == "no rows in result set" {
			log.Info("admin account already exists", "email", email)
			return nil
		}
		return fmt.Errorf("insert admin: %w", err)
	}
	log.Info("admin account created", "email", email, "id", id)
	return nil
}
