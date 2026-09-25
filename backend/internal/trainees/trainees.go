// Package trainees implements staff-facing trainee management: manual create,
// list/detail/update, account status, and validated CSV import.
package trainees

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/skycode/ojt-management/backend/internal/auth"
	"github.com/skycode/ojt-management/backend/internal/platform/db"
	"github.com/skycode/ojt-management/backend/internal/platform/httpx"
)

type Trainee struct {
	ID              uuid.UUID  `json:"id"`
	UserID          uuid.UUID  `json:"user_id"`
	Email           string     `json:"email"`
	DisplayName     string     `json:"display_name"`
	StudentNumber   string     `json:"student_number"`
	Program         string     `json:"program"`
	YearLevel       string     `json:"year_level"`
	ContactNumber   string     `json:"contact_number"`
	AccountStatus   string     `json:"account_status"`
	SiteName        *string    `json:"site_name,omitempty"`
	AssignmentID    *uuid.UUID `json:"assignment_id,omitempty"`
	ProgressPercent *float64   `json:"progress_percent,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type CreateInput struct {
	Email         string `json:"email"`
	DisplayName   string `json:"display_name"`
	StudentNumber string `json:"student_number"`
	Program       string `json:"program"`
	YearLevel     string `json:"year_level"`
	ContactNumber string `json:"contact_number"`
}

type PatchInput struct {
	DisplayName   *string `json:"display_name"`
	StudentNumber *string `json:"student_number"`
	Program       *string `json:"program"`
	YearLevel     *string `json:"year_level"`
	ContactNumber *string `json:"contact_number"`
	AccountStatus *string `json:"account_status"`
}

// TempPassword generates a one-time bootstrap password shown to staff once.
func TempPassword() (string, error) {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func validateCreate(in *CreateInput, fields map[string]string) {
	if strings.TrimSpace(in.Email) == "" || !strings.Contains(in.Email, "@") {
		fields["email"] = "A valid email is required."
	}
	if strings.TrimSpace(in.DisplayName) == "" {
		fields["display_name"] = "Display name is required."
	}
	if strings.TrimSpace(in.StudentNumber) == "" {
		fields["student_number"] = "Student number is required."
	}
}

const traineeSelect = `
	SELECT tp.id, tp.user_id, u.email, u.display_name, tp.student_number,
	       COALESCE(tp.program,''), COALESCE(tp.year_level,''), COALESCE(tp.contact_number,''),
	       u.account_status, s.name, a.id, tp.created_at
	FROM trainee_profiles tp
	JOIN users u ON u.id = tp.user_id
	LEFT JOIN ojt_assignments a ON a.trainee_id = tp.id AND a.status = 'active'
	LEFT JOIN ojt_sites s ON s.id = a.site_id `

func scanTrainee(row pgx.Row) (*Trainee, error) {
	var t Trainee
	err := row.Scan(&t.ID, &t.UserID, &t.Email, &t.DisplayName, &t.StudentNumber,
		&t.Program, &t.YearLevel, &t.ContactNumber, &t.AccountStatus,
		&t.SiteName, &t.AssignmentID, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func scanTraineeRows(rows pgx.Rows) ([]Trainee, error) {
	defer rows.Close()
	var out []Trainee
	for rows.Next() {
		var t Trainee
		if err := rows.Scan(&t.ID, &t.UserID, &t.Email, &t.DisplayName, &t.StudentNumber,
			&t.Program, &t.YearLevel, &t.ContactNumber, &t.AccountStatus,
			&t.SiteName, &t.AssignmentID, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func Get(ctx context.Context, pool *db.Pool, id uuid.UUID) (*Trainee, error) {
	t, err := scanTrainee(pool.QueryRow(ctx, traineeSelect+` WHERE tp.id = $1`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httpx.ErrNotFound
		}
		return nil, httpx.Internal(err)
	}
	return t, nil
}

func GetByUserID(ctx context.Context, pool *db.Pool, userID uuid.UUID) (*Trainee, error) {
	t, err := scanTrainee(pool.QueryRow(ctx, traineeSelect+` WHERE tp.user_id = $1`, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return t, nil
}

// List returns a scope-filtered paginated trainee list.
func List(ctx context.Context, pool *db.Pool, p httpx.Page, scopeIDs []uuid.UUID, q, status, siteID string) ([]Trainee, int64, error) {
	where := `WHERE ($1::uuid[] IS NULL OR tp.id = ANY($1))
		AND ($2::text IS NULL OR u.display_name ILIKE '%'||$2||'%' OR u.email ILIKE '%'||$2||'%' OR tp.student_number ILIKE '%'||$2||'%')
		AND ($3::text IS NULL OR u.account_status = $3)
		AND ($4::uuid IS NULL OR a.site_id = $4)`
	var scopeArg, qArg, statusArg, siteArg any
	if scopeIDs != nil {
		scopeArg = scopeIDs
	}
	if q = strings.TrimSpace(q); q != "" {
		qArg = q
	}
	if status == "active" || status == "inactive" || status == "locked" {
		statusArg = status
	}
	if v, err := uuid.Parse(siteID); err == nil {
		siteArg = v
	}

	var total int64
	err := pool.QueryRow(ctx,
		`SELECT count(*) FROM trainee_profiles tp
		 JOIN users u ON u.id = tp.user_id
		 LEFT JOIN ojt_assignments a ON a.trainee_id = tp.id AND a.status = 'active'
		 `+where, scopeArg, qArg, statusArg, siteArg).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := pool.Query(ctx, traineeSelect+where+`
		ORDER BY u.display_name LIMIT $5 OFFSET $6`,
		scopeArg, qArg, statusArg, siteArg, p.Limit(), p.Offset())
	if err != nil {
		return nil, 0, err
	}
	out, err := scanTraineeRows(rows)
	return out, total, err
}

// Create inserts user + trainee profile atomically. When the creator is a
// coordinator the trainee is added to their scope. Returns the created
// trainee and a one-time temporary password.
func Create(ctx context.Context, pool *db.Pool, in CreateInput, creator *auth.User) (*Trainee, string, error) {
	fields := map[string]string{}
	validateCreate(&in, fields)
	if len(fields) > 0 {
		return nil, "", httpx.Validation(fields)
	}

	tempPassword, err := TempPassword()
	if err != nil {
		return nil, "", httpx.Internal(err)
	}
	hash, err := auth.HashPassword(tempPassword)
	if err != nil {
		return nil, "", httpx.Internal(err)
	}

	var traineeID uuid.UUID
	err = pool.InTx(ctx, func(tx pgx.Tx) error {
		var userID uuid.UUID
		err := tx.QueryRow(ctx, `
			INSERT INTO users (email, password_hash, role, display_name)
			VALUES ($1, $2, 'trainee', $3) RETURNING id`,
			strings.ToLower(strings.TrimSpace(in.Email)), hash, strings.TrimSpace(in.DisplayName)).Scan(&userID)
		if err != nil {
			if db.IsUniqueViolation(err, "users_email_lower_key") {
				return httpx.Validation(map[string]string{"email": "A user with this email already exists."})
			}
			return err
		}
		err = tx.QueryRow(ctx, `
			INSERT INTO trainee_profiles (user_id, student_number, program, year_level, contact_number)
			VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			userID, strings.TrimSpace(in.StudentNumber), nullIfEmpty(in.Program),
			nullIfEmpty(in.YearLevel), nullIfEmpty(in.ContactNumber)).Scan(&traineeID)
		if err != nil {
			if db.IsUniqueViolation(err, "trainee_profiles_student_number_key") {
				return httpx.Validation(map[string]string{"student_number": "Student number already exists."})
			}
			return err
		}
		if creator.Role == "coordinator" {
			_, err = tx.Exec(ctx, `
				INSERT INTO coordinator_scopes (coordinator_user_id, trainee_id) VALUES ($1, $2)
				ON CONFLICT DO NOTHING`, creator.ID, traineeID)
		}
		return err
	})
	if err != nil {
		var ae *httpx.APIError
		if errors.As(err, &ae) {
			return nil, "", ae
		}
		return nil, "", httpx.Internal(err)
	}
	t, err := Get(ctx, pool, traineeID)
	return t, tempPassword, err
}

// Update applies profile/account-state changes. Setting account_status to a
// non-active value revokes all sessions for the user.
func Update(ctx context.Context, pool *db.Pool, id uuid.UUID, in PatchInput) (*Trainee, error) {
	cur, err := Get(ctx, pool, id)
	if err != nil {
		return nil, err
	}
	name := cur.DisplayName
	if in.DisplayName != nil {
		name = strings.TrimSpace(*in.DisplayName)
	}
	if name == "" {
		return nil, httpx.Validation(map[string]string{"display_name": "Display name is required."})
	}
	studentNo := cur.StudentNumber
	if in.StudentNumber != nil {
		studentNo = strings.TrimSpace(*in.StudentNumber)
	}
	if studentNo == "" {
		return nil, httpx.Validation(map[string]string{"student_number": "Student number is required."})
	}
	status := cur.AccountStatus
	if in.AccountStatus != nil {
		if *in.AccountStatus != "active" && *in.AccountStatus != "inactive" && *in.AccountStatus != "locked" {
			return nil, httpx.Validation(map[string]string{"account_status": "Must be active, inactive, or locked."})
		}
		status = *in.AccountStatus
	}
	program, year, contact := cur.Program, cur.YearLevel, cur.ContactNumber
	if in.Program != nil {
		program = *in.Program
	}
	if in.YearLevel != nil {
		year = *in.YearLevel
	}
	if in.ContactNumber != nil {
		contact = *in.ContactNumber
	}

	err = pool.InTx(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx,
			`UPDATE users SET display_name = $2, account_status = $3, updated_at = now() WHERE id = $1`,
			cur.UserID, name, status); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			UPDATE trainee_profiles SET student_number = $2, program = $3, year_level = $4,
				contact_number = $5, updated_at = now() WHERE id = $1`,
			id, studentNo, nullIfEmpty(program), nullIfEmpty(year), nullIfEmpty(contact)); err != nil {
			return err
		}
		if status != "active" && status != cur.AccountStatus {
			// Revoke all live sessions on deactivation.
			if _, err := tx.Exec(ctx,
				`UPDATE user_sessions SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL`,
				cur.UserID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		if db.IsUniqueViolation(err, "trainee_profiles_student_number_key") {
			return nil, httpx.Validation(map[string]string{"student_number": "Student number already exists."})
		}
		return nil, httpx.Internal(err)
	}
	return Get(ctx, pool, id)
}

// SetPassword replaces a trainee account password and signs out any active
// trainee sessions. Scope checks are performed by the HTTP handler first.
func SetPassword(ctx context.Context, pool *db.Pool, id uuid.UUID, newPassword string) error {
	if len(newPassword) < 8 {
		return httpx.Validation(map[string]string{"new_password": "Password must be at least 8 characters."})
	}
	t, err := Get(ctx, pool, id)
	if err != nil {
		return err
	}
	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		return httpx.Internal(err)
	}
	if err := pool.InTx(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx,
			`UPDATE users SET password_hash = $2, updated_at = now() WHERE id = $1`,
			t.UserID, hash); err != nil {
			return err
		}
		_, err := tx.Exec(ctx,
			`UPDATE user_sessions SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL`,
			t.UserID)
		return err
	}); err != nil {
		return httpx.Internal(err)
	}
	return nil
}

func nullIfEmpty(s string) any {
	if s = strings.TrimSpace(s); s != "" {
		return s
	}
	return nil
}
