// Authentication middleware, role guards, CSRF enforcement, and the
// coordinator-scope authorization helper.
package auth

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/skycode/ojt-management/backend/internal/platform/db"
	"github.com/skycode/ojt-management/backend/internal/platform/httpx"
)

type ctxKey string

const (
	ctxUser    ctxKey = "auth.user"
	ctxSession ctxKey = "auth.session"
	ctxToken   ctxKey = "auth.token"
)

// ActorOf returns the authenticated user, or nil.
func ActorOf(c fiber.Ctx) *User {
	if u, ok := c.Locals(ctxUser).(*User); ok {
		return u
	}
	return nil
}

func SessionOf(c fiber.Ctx) *Session {
	if s, ok := c.Locals(ctxSession).(*Session); ok {
		return s
	}
	return nil
}

func TokenOf(c fiber.Ctx) string {
	if t, ok := c.Locals(ctxToken).(string); ok {
		return t
	}
	return ""
}

// RequireAuth resolves the session cookie into request-local actor state.
func RequireAuth(sm *SessionManager) fiber.Handler {
	return func(c fiber.Ctx) error {
		token := c.Cookies(sm.cfg.CookieName)
		if token == "" {
			return httpx.Fail(c, httpx.ErrUnauthorized)
		}
		sess, user, err := sm.Resolve(c.Context(), token)
		if err != nil {
			return httpx.Fail(c, httpx.ErrUnauthorized)
		}
		c.Locals(ctxUser, user)
		c.Locals(ctxSession, sess)
		c.Locals(ctxToken, token)
		return c.Next()
	}
}

// CSRFProtect rejects unsafe requests that lack a valid per-session CSRF
// token. It runs after RequireAuth. The Origin header, when present, must be
// in the allowed-origins set (same-origin deployments send an allowed origin).
func CSRFProtect(allowedOrigins []string, sm *SessionManager) fiber.Handler {
	allowed := map[string]bool{}
	for _, o := range allowedOrigins {
		allowed[o] = true
	}
	return func(c fiber.Ctx) error {
		switch c.Method() {
		case fiber.MethodGet, fiber.MethodHead, fiber.MethodOptions:
			return c.Next()
		}
		if origin := c.Get(fiber.HeaderOrigin); origin != "" && !allowed[origin] {
			return httpx.Fail(c, httpx.ErrCSRF)
		}
		if !sm.ValidCSRF(SessionOf(c), c.Get("X-CSRF-Token")) {
			return httpx.Fail(c, httpx.ErrCSRF)
		}
		return c.Next()
	}
}

// RequireRole allows the listed roles only.
func RequireRole(roles ...string) fiber.Handler {
	set := map[string]bool{}
	for _, r := range roles {
		set[r] = true
	}
	return func(c fiber.Ctx) error {
		u := ActorOf(c)
		if u == nil {
			return httpx.Fail(c, httpx.ErrUnauthorized)
		}
		if !set[u.Role] {
			return httpx.Fail(c, httpx.ErrForbidden)
		}
		return c.Next()
	}
}

// ---- Coordinator scope ----------------------------------------------------

// IsStaff reports whether the user is coordinator or admin.
func IsStaff(u *User) bool { return u.Role == "coordinator" || u.Role == "admin" }

// CanAccessTrainee enforces scope rules:
//   - admin: institution-wide
//   - coordinator: trainee must appear in coordinator_scopes
//   - trainee: only themselves (caller compares ownership)
func CanAccessTrainee(ctx context.Context, pool *db.Pool, staff *User, traineeID uuid.UUID) (bool, error) {
	if staff.Role == "admin" {
		return true, nil
	}
	if staff.Role != "coordinator" {
		return false, nil
	}
	var exists bool
	err := pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM coordinator_scopes WHERE coordinator_user_id = $1 AND trainee_id = $2)`,
		staff.ID, traineeID).Scan(&exists)
	return exists, err
}

// ScopedTraineeIDs returns nil for admin (meaning: no filter) and the list of
// scoped trainee IDs for a coordinator. Callers use:
//
//	WHERE ($1::uuid[] IS NULL OR trainee_id = ANY($1))
//
// An empty non-nil slice means the coordinator has no scope → matches nothing.
func ScopedTraineeIDs(ctx context.Context, pool *db.Pool, staff *User) ([]uuid.UUID, error) {
	if staff.Role == "admin" {
		return nil, nil
	}
	if staff.Role != "coordinator" {
		return []uuid.UUID{}, nil
	}
	rows, err := pool.Query(ctx,
		`SELECT trainee_id FROM coordinator_scopes WHERE coordinator_user_id = $1`, staff.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []uuid.UUID{}
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
