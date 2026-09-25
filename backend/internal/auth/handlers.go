// HTTP handlers for /api/v1/auth/*.
package auth

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/skycode/ojt-management/backend/internal/audit"
	"github.com/skycode/ojt-management/backend/internal/config"
	"github.com/skycode/ojt-management/backend/internal/platform/db"
	"github.com/skycode/ojt-management/backend/internal/platform/httpx"
	"github.com/skycode/ojt-management/backend/internal/platform/ratelimit"
)

type Handler struct {
	svc  *Service
	cfg  config.AuthConfig
	sm   *SessionManager
	pool *db.Pool
}

func NewHandler(svc *Service, sm *SessionManager, pool *db.Pool, cfg config.AuthConfig) *Handler {
	return &Handler{svc: svc, cfg: cfg, sm: sm, pool: pool}
}

type userDTO struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
	Email       string `json:"email"`
}

func toUserDTO(u *User) userDTO {
	return userDTO{ID: u.ID.String(), DisplayName: u.DisplayName, Role: u.Role, Email: u.Email}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// Login: POST /api/v1/auth/login — public, rate-limited per IP.
func (h *Handler) Login(c fiber.Ctx) error {
	var req loginRequest
	if err := c.Bind().JSON(&req); err != nil {
		return httpx.Fail(c, httpx.ValidationMsg("Malformed request body."))
	}
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		return httpx.Fail(c, httpx.Validation(map[string]string{
			"email":    "Email is required.",
			"password": "Password is required.",
		}))
	}

	res, err := h.svc.Login(c.Context(), req.Email, req.Password, c.Get("User-Agent"), c.IP())
	if err != nil {
		audit.Record(c.Context(), h.pool, nil, "auth.login_failed", "user", nil, httpx.RequestID(c), nil)
		return httpx.Fail(c, err)
	}

	h.setSessionCookie(c, res.Token, res.ExpiresAt)
	audit.Record(c.Context(), h.pool, &res.User.ID, "auth.login", "user", &res.User.ID, httpx.RequestID(c), nil)

	return httpx.OK(c, fiber.Map{
		"user":       toUserDTO(&res.User),
		"csrf_token": res.CSRFToken,
	})
}

// Logout: POST /api/v1/auth/logout — revokes the session; idempotent.
func (h *Handler) Logout(c fiber.Ctx) error {
	token := TokenOf(c)
	if token == "" {
		token = c.Cookies(h.cfg.CookieName)
	}
	if err := h.svc.Logout(c.Context(), token); err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	if u := ActorOf(c); u != nil {
		audit.Record(c.Context(), h.pool, &u.ID, "auth.logout", "user", &u.ID, httpx.RequestID(c), nil)
	}
	h.clearSessionCookie(c)
	return httpx.OK(c, fiber.Map{"logged_out": true})
}

// Me: GET /api/v1/auth/me — current user context.
func (h *Handler) Me(c fiber.Ctx) error {
	u := ActorOf(c)
	if u == nil {
		return httpx.Fail(c, httpx.ErrUnauthorized)
	}
	return httpx.OK(c, fiber.Map{"user": toUserDTO(u)})
}

// CSRFToken: GET /api/v1/auth/csrf — returns a fresh CSRF token bound to the
// current session and rotates the stored secret hash.
func (h *Handler) CSRFToken(c fiber.Ctx) error {
	sess := SessionOf(c)
	if sess == nil {
		return httpx.Fail(c, httpx.ErrUnauthorized)
	}
	token, err := h.sm.RotateCSRF(c.Context(), sess.ID)
	if err != nil {
		return httpx.Fail(c, httpx.Internal(err))
	}
	return httpx.OK(c, fiber.Map{"csrf_token": token})
}

// ChangePassword: POST /api/v1/auth/change-password — authenticated users
// change their own password after confirming the current password.
func (h *Handler) ChangePassword(c fiber.Ctx) error {
	var req changePasswordRequest
	if err := c.Bind().JSON(&req); err != nil {
		return httpx.Fail(c, httpx.ValidationMsg("Malformed request body."))
	}
	u := ActorOf(c)
	if err := h.svc.ChangePassword(c.Context(), u, req.CurrentPassword, req.NewPassword); err != nil {
		return httpx.Fail(c, err)
	}
	audit.Record(c.Context(), h.pool, &u.ID, "auth.password_changed", "user", &u.ID, httpx.RequestID(c), nil)
	return httpx.OK(c, fiber.Map{"updated": true})
}

func (h *Handler) setSessionCookie(c fiber.Ctx, token string, expires time.Time) {
	c.Cookie(&fiber.Cookie{
		Name:     h.cfg.CookieName,
		Value:    token,
		Path:     "/",
		Domain:   h.cfg.CookieDomain,
		Expires:  expires,
		Secure:   h.cfg.CookieSecure,
		HTTPOnly: true,
		SameSite: "Lax",
	})
}

func (h *Handler) clearSessionCookie(c fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     h.cfg.CookieName,
		Value:    "",
		Path:     "/",
		Domain:   h.cfg.CookieDomain,
		MaxAge:   -1,
		Secure:   h.cfg.CookieSecure,
		HTTPOnly: true,
		SameSite: "Lax",
	})
}

// RegisterRoutes wires auth routes. The login endpoint is rate-limited;
// logout/me/csrf require auth; logout also passes CSRF protection.
func RegisterRoutes(api fiber.Router, h *Handler, sm *SessionManager, cfg config.AuthConfig) {
	loginLimiter := ratelimit.New(cfg.LoginRateLimit, cfg.LoginRateWindow)

	auth := api.Group("/auth")
	auth.Post("/login", loginLimiter.Middleware(), h.Login)
	auth.Post("/logout", RequireAuth(sm), CSRFProtect(cfg.AllowedOrigins, sm), h.Logout)
	auth.Post("/change-password", RequireAuth(sm), CSRFProtect(cfg.AllowedOrigins, sm), h.ChangePassword)
	auth.Get("/me", RequireAuth(sm), h.Me)
	auth.Get("/csrf", RequireAuth(sm), h.CSRFToken)
}
