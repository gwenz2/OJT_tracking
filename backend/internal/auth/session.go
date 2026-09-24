// Opaque server-side sessions. The client receives a random bearer token in a
// Secure+HttpOnly+SameSite cookie; only its SHA-256 hash is stored. A
// per-session CSRF secret (also stored hashed) protects unsafe requests.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/skycode/ojt-management/backend/internal/config"
	"github.com/skycode/ojt-management/backend/internal/platform/db"
)

var ErrSessionInvalid = errors.New("session is invalid or expired")

type Session struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ExpiresAt time.Time
	CSRFHash  string
}

// User is the authenticated principal resolved from a session.
type User struct {
	ID          uuid.UUID
	Email       string
	Role        string
	Status      string
	DisplayName string
}

type SessionManager struct {
	pool *db.Pool
	cfg  config.AuthConfig
}

func NewSessionManager(pool *db.Pool, cfg config.AuthConfig) *SessionManager {
	return &SessionManager{pool: pool, cfg: cfg}
}

func randToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// Create issues a new session and returns the bearer token and CSRF token.
// uaHash and ip are minimized audit metadata (user-agent hash, masked IP).
func (m *SessionManager) Create(ctx context.Context, userID uuid.UUID, userAgent, remoteIP string) (token, csrfToken string, err error) {
	token, err = randToken()
	if err != nil {
		return "", "", err
	}
	csrfToken, err = randToken()
	if err != nil {
		return "", "", err
	}

	var ipPrefix *net.IP
	if ip := net.ParseIP(remoteIP); ip != nil {
		masked := maskIP(ip)
		ipPrefix = &masked
	}

	_, err = m.pool.Exec(ctx, `
		INSERT INTO user_sessions (user_id, token_hash, csrf_secret_hash, user_agent_hash, ip_prefix, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		userID, sha256Hex(token), sha256Hex(csrfToken), sha256Hex(userAgent), ipPrefix,
		time.Now().UTC().Add(m.cfg.SessionTTL))
	if err != nil {
		return "", "", err
	}
	return token, csrfToken, nil
}

// maskIP truncates the address for minimized audit storage (/24 v4, /48 v6).
func maskIP(ip net.IP) net.IP {
	if v4 := ip.To4(); v4 != nil {
		v4[3] = 0
		return v4
	}
	v6 := ip.To16()
	for i := 6; i < 16; i++ {
		v6[i] = 0
	}
	return v6
}

// Resolve looks up a session token and returns the session and user when the
// session is unexpired, unrevoked, and the account is usable.
func (m *SessionManager) Resolve(ctx context.Context, token string) (*Session, *User, error) {
	var (
		sess Session
		u    User
	)
	err := m.pool.QueryRow(ctx, `
		SELECT s.id, s.user_id, s.expires_at, COALESCE(s.csrf_secret_hash, ''),
		       u.id, u.email, u.role, u.account_status, u.display_name
		FROM user_sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1
		  AND s.revoked_at IS NULL
		  AND s.expires_at > now()`, sha256Hex(token)).
		Scan(&sess.ID, &sess.UserID, &sess.ExpiresAt, &sess.CSRFHash,
			&u.ID, &u.Email, &u.Role, &u.Status, &u.DisplayName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, ErrSessionInvalid
		}
		return nil, nil, err
	}
	if u.Status != "active" {
		return nil, nil, ErrSessionInvalid
	}

	// Best-effort last_seen update, throttled by the DB write being trivially
	// small; failure does not break the request.
	_, _ = m.pool.Exec(ctx,
		`UPDATE user_sessions SET last_seen_at = now() WHERE id = $1`, sess.ID)

	return &sess, &u, nil
}

// Revoke marks the session identified by token revoked. Idempotent.
func (m *SessionManager) Revoke(ctx context.Context, token string) error {
	_, err := m.pool.Exec(ctx,
		`UPDATE user_sessions SET revoked_at = now() WHERE token_hash = $1 AND revoked_at IS NULL`,
		sha256Hex(token))
	return err
}

// RevokeAllForUser revokes every live session for a user (account disable).
func (m *SessionManager) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	_, err := m.pool.Exec(ctx,
		`UPDATE user_sessions SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL`,
		userID)
	return err
}

// RotateCSRF generates a new CSRF token for an existing session and returns
// it. The stored hash is replaced, so the returned token must reach the client.
func (m *SessionManager) RotateCSRF(ctx context.Context, sessionID uuid.UUID) (string, error) {
	token, err := randToken()
	if err != nil {
		return "", err
	}
	_, err = m.pool.Exec(ctx,
		`UPDATE user_sessions SET csrf_secret_hash = $2 WHERE id = $1`,
		sessionID, sha256Hex(token))
	if err != nil {
		return "", err
	}
	return token, nil
}

// ValidCSRF reports whether the presented CSRF token matches the session's
// stored secret hash, in constant time.
func (m *SessionManager) ValidCSRF(sess *Session, presented string) bool {
	if sess == nil || presented == "" || sess.CSRFHash == "" {
		return false
	}
	want, _ := hex.DecodeString(sess.CSRFHash)
	got := sha256.Sum256([]byte(presented))
	return subtle.ConstantTimeCompare(got[:], want) == 1
}
