// Login/logout use-cases.
package auth

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/skycode/ojt-management/backend/internal/config"
	"github.com/skycode/ojt-management/backend/internal/platform/db"
	"github.com/skycode/ojt-management/backend/internal/platform/httpx"
)

var (
	ErrInvalidCredentials = httpx.New(401, "INVALID_CREDENTIALS", "Invalid email or password.")
	ErrAccountInactive    = httpx.New(403, "ACCOUNT_INACTIVE", "This account is not active. Contact your coordinator.")
)

type Service struct {
	pool *db.Pool
	sm   *SessionManager
	cfg  config.AuthConfig
}

func NewService(pool *db.Pool, sm *SessionManager, cfg config.AuthConfig) *Service {
	return &Service{pool: pool, sm: sm, cfg: cfg}
}

// LoginResult carries everything the handler needs to complete login.
type LoginResult struct {
	User      User
	Token     string
	CSRFToken string
	ExpiresAt time.Time
}

// Login verifies credentials and creates a session. The credential-failure
// path is deliberately generic and takes a similar amount of time whether the
// email exists or not.
func (s *Service) Login(ctx context.Context, email, password, userAgent, ip string) (*LoginResult, error) {
	var u User
	var hash string
	err := s.pool.QueryRow(ctx, `
		SELECT id, email, role, account_status, display_name, password_hash
		FROM users WHERE lower(email) = lower($1)`, email).
		Scan(&u.ID, &u.Email, &u.Role, &u.Status, &u.DisplayName, &hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Verify against a dummy hash so the timing does not reveal
			// whether the account exists.
			_, _, _ = VerifyPassword(password, dummyHash)
			return nil, ErrInvalidCredentials
		}
		return nil, httpx.Internal(err)
	}

	match, needsRehash, err := VerifyPassword(password, hash)
	if err != nil || !match {
		return nil, ErrInvalidCredentials
	}
	if u.Status != "active" {
		return nil, ErrAccountInactive
	}
	if needsRehash {
		if newHash, err := HashPassword(password); err == nil {
			_, _ = s.pool.Exec(ctx, `UPDATE users SET password_hash = $2 WHERE id = $1`, u.ID, newHash)
		}
	}

	token, csrf, err := s.sm.Create(ctx, u.ID, userAgent, ip)
	if err != nil {
		return nil, httpx.Internal(err)
	}
	_, _ = s.pool.Exec(ctx, `UPDATE users SET last_login_at = now() WHERE id = $1`, u.ID)

	return &LoginResult{
		User:      u,
		Token:     token,
		CSRFToken: csrf,
		ExpiresAt: time.Now().UTC().Add(s.cfg.SessionTTL),
	}, nil
}

// Logout revokes the presented session token. Idempotent.
func (s *Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.sm.Revoke(ctx, token)
}

// ChangePassword verifies the current password before replacing the caller's
// password hash. The active session remains valid.
func (s *Service) ChangePassword(ctx context.Context, u *User, currentPassword, newPassword string) error {
	if u == nil {
		return httpx.ErrUnauthorized
	}
	fields := map[string]string{}
	if currentPassword == "" {
		fields["current_password"] = "Current password is required."
	}
	if len(newPassword) < 8 {
		fields["new_password"] = "New password must be at least 8 characters."
	}
	if len(fields) > 0 {
		return httpx.Validation(fields)
	}

	var hash string
	err := s.pool.QueryRow(ctx, `SELECT password_hash FROM users WHERE id = $1`, u.ID).Scan(&hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrUnauthorized
		}
		return httpx.Internal(err)
	}
	match, _, err := VerifyPassword(currentPassword, hash)
	if err != nil || !match {
		return httpx.Validation(map[string]string{"current_password": "Current password is incorrect."})
	}
	newHash, err := HashPassword(newPassword)
	if err != nil {
		return httpx.Internal(err)
	}
	if _, err := s.pool.Exec(ctx,
		`UPDATE users SET password_hash = $2, updated_at = now() WHERE id = $1`,
		u.ID, newHash); err != nil {
		return httpx.Internal(err)
	}
	return nil
}

// dummyHash is a valid-format Argon2id hash used to equalize login timing when
// the email does not exist. It never matches a real password.
const dummyHash = "$argon2id$v=19$m=65536,t=2,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
