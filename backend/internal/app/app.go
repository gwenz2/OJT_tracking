// Package app wires the Fiber application: middleware, health, and every
// domain module's routes. Shared by cmd/api and the contract/integration
// tests.
package app

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"

	"github.com/skycode/ojt-management/backend/internal/assignments"
	"github.com/skycode/ojt-management/backend/internal/attendance"
	"github.com/skycode/ojt-management/backend/internal/auth"
	"github.com/skycode/ojt-management/backend/internal/config"
	"github.com/skycode/ojt-management/backend/internal/corrections"
	"github.com/skycode/ojt-management/backend/internal/journals"
	"github.com/skycode/ojt-management/backend/internal/monitoring"
	"github.com/skycode/ojt-management/backend/internal/notifications"
	"github.com/skycode/ojt-management/backend/internal/platform/db"
	"github.com/skycode/ojt-management/backend/internal/platform/httpx"
	"github.com/skycode/ojt-management/backend/internal/platform/observability"
	"github.com/skycode/ojt-management/backend/internal/platform/ratelimit"
	"github.com/skycode/ojt-management/backend/internal/platform/storage"
	"github.com/skycode/ojt-management/backend/internal/reports"
	"github.com/skycode/ojt-management/backend/internal/settings"
	"github.com/skycode/ojt-management/backend/internal/sites"
	"github.com/skycode/ojt-management/backend/internal/trainees"
	"github.com/skycode/ojt-management/backend/internal/users"
)

// Deps carries everything the app needs that is constructed at startup.
type Deps struct {
	Cfg  *config.Config
	Log  *slog.Logger
	Pool *db.Pool
	// Store may be nil in tests that do not touch evidence paths.
	Store *storage.Store
}

func Build(d Deps) *fiber.App {
	cfg := d.Cfg
	app := fiber.New(fiber.Config{
		AppName:      cfg.AppName,
		ErrorHandler: httpx.ErrorHandler,
		BodyLimit:    int(cfg.EvidenceUploadLimitBytes),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	app.Use(recover.New())
	app.Use(httpx.RequestIDMiddleware())
	app.Use(observability.RequestLogger(d.Log))

	healthDeps := httpx.HealthDeps{CheckDB: d.Pool.Health}
	if d.Store != nil {
		healthDeps.CheckStorage = d.Store.Health
	}
	httpx.RegisterHealth(app, healthDeps)

	api := app.Group("/api/v1")

	sm := auth.NewSessionManager(d.Pool, cfg.Auth)
	authSvc := auth.NewService(d.Pool, sm, cfg.Auth)
	authHandler := auth.NewHandler(authSvc, sm, d.Pool, cfg.Auth)
	auth.RegisterRoutes(api, authHandler, sm, cfg.Auth)

	csrf := auth.CSRFProtect(cfg.Auth.AllowedOrigins, sm)
	staff := api.Group("/staff",
		auth.RequireAuth(sm), csrf, auth.RequireRole("coordinator", "admin"))
	admin := api.Group("/admin",
		auth.RequireAuth(sm), csrf, auth.RequireRole("admin"))
	trainee := api.Group("/trainee",
		auth.RequireAuth(sm), csrf, auth.RequireRole("trainee"))

	trainee.Get("/dashboard", trainees.NewDashboardHandler(d.Pool).Get)

	// Attendance: writes are trainee-only and per-user rate limited; reads are
	// authorized per-handler (owner trainee or in-scope staff).
	attendanceSvc := attendance.NewService(d.Pool, d.Store, cfg.EvidenceUploadLimitBytes)
	attendanceH := attendance.NewHandler(attendanceSvc)
	attendanceRL := ratelimit.New(cfg.Auth.AttendanceRateLimit, cfg.Auth.AttendanceRateWindow)
	att := api.Group("/attendance", auth.RequireAuth(sm), csrf)
	att.Post("/time-in",
		auth.RequireRole("trainee"),
		attendanceRL.KeyedMiddleware(func(c fiber.Ctx) string {
			if u := auth.ActorOf(c); u != nil {
				return "attendance:" + u.ID.String()
			}
			return ""
		}),
		attendanceH.TimeIn)
	att.Post("/:id/time-out",
		auth.RequireRole("trainee"),
		attendanceRL.KeyedMiddleware(func(c fiber.Ctx) string {
			if u := auth.ActorOf(c); u != nil {
				return "attendance:" + u.ID.String()
			}
			return ""
		}),
		attendanceH.TimeOut)

	readH := attendance.NewReadHandler(d.Pool, d.Store)
	att.Get("/today", auth.RequireRole("trainee"), readH.Today)
	att.Get("/history", auth.RequireRole("trainee"), readH.History)
	att.Get("/:id", readH.Detail)
	att.Get("/:id/evidence/:action/image", readH.EvidenceImage)

	// Journals: shared read for owner/staff; trainee-only mutations.
	journalsH := journals.NewHandler(d.Pool, d.Store, cfg.EvidenceUploadLimitBytes)
	jr := api.Group("/journals", auth.RequireAuth(sm), csrf)
	jr.Get("/:id", journalsH.Get)
	jr.Put("/:id/draft", auth.RequireRole("trainee"), journalsH.SaveDraft)
	jr.Post("/:id/evidence", auth.RequireRole("trainee"), journalsH.AddEvidence)
	jr.Delete("/:id/evidence/:eid", auth.RequireRole("trainee"), journalsH.DeleteEvidence)
	jr.Post("/:id/submit", auth.RequireRole("trainee"), journalsH.Submit)
	jr.Get("/:id/evidence/:eid/image", journalsH.EvidenceImage)

	staff.Get("/journals", journalsH.Queue)
	staff.Post("/journals/:id/review", journalsH.Review)

	// Corrections: trainee proposals + staff decisions.
	correctionsH := corrections.NewHandler(d.Pool)
	att.Post("/:id/corrections",
		auth.RequireRole("trainee"),
		attendanceRL.KeyedMiddleware(func(c fiber.Ctx) string {
			if u := auth.ActorOf(c); u != nil {
				return "correction:" + u.ID.String()
			}
			return ""
		}),
		correctionsH.Create)
	trainee.Get("/corrections", correctionsH.Mine)
	staff.Get("/corrections", correctionsH.List)
	staff.Post("/corrections/:id/decision", correctionsH.Decide)

	// Staff monitoring: dashboard metrics + attendance list.
	monH := monitoring.NewHandler(d.Pool)
	staff.Get("/dashboard", monH.Dashboard)
	staff.Get("/attendance", monH.AttendanceList)

	// Staff reports: JSON rows + CSV export.
	reportsH := reports.NewHandler(d.Pool)
	staff.Get("/reports/:name/export.csv", reportsH.Export)
	staff.Get("/reports/:name", reportsH.Get)

	// Notifications: any authenticated user reads their own.
	notifH := notifications.NewHandler(d.Pool)
	ntf := api.Group("/notifications", auth.RequireAuth(sm), csrf)
	ntf.Get("/", notifH.List)
	ntf.Post("/:id/read", notifH.Read)
	ntf.Post("/read-all", notifH.ReadAll)

	sitesH := sites.NewHandler(d.Pool)
	staff.Get("/sites", sitesH.List)
	staff.Post("/sites", sitesH.Create)
	staff.Get("/sites/:id", sitesH.GetOne)
	staff.Patch("/sites/:id", sitesH.Update)

	assignmentsH := assignments.NewHandler(d.Pool)
	staff.Get("/assignments", assignmentsH.List)
	staff.Post("/assignments", assignmentsH.Create)
	staff.Get("/assignments/:id", assignmentsH.GetOne)
	staff.Patch("/assignments/:id", assignmentsH.Update)

	traineesH := trainees.NewHandler(d.Pool)
	staff.Get("/trainees", traineesH.List)
	staff.Post("/trainees", traineesH.Create)
	staff.Post("/trainees/import/validate", traineesH.ImportValidate)
	staff.Post("/trainees/import/commit", traineesH.ImportCommit)
	staff.Get("/trainees/:id", traineesH.GetOne)
	staff.Patch("/trainees/:id", traineesH.Update)
	staff.Post("/trainees/:id/password", traineesH.SetPassword)

	usersH := users.NewHandler(d.Pool)
	admin.Get("/coordinators", usersH.List)
	admin.Post("/coordinators", usersH.Create)
	admin.Patch("/coordinators/:id", usersH.Patch)
	admin.Put("/coordinators/:id/trainee-scope", usersH.SetScope)
	admin.Post("/coordinators/:id/password", usersH.SetPassword)

	settingsH := settings.NewHandler(d.Pool)
	admin.Get("/settings", settingsH.Get)
	admin.Patch("/settings", settingsH.Patch)

	return app
}
