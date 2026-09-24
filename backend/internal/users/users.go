// Package users implements administrator management of coordinator accounts
// and coordinator→trainee scope assignment.
package users

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/skycode/ojt-management/backend/internal/auth"
	"github.com/skycode/ojt-management/backend/internal/platform/db"
	"github.com/skycode/ojt-management/backend/internal/platform/httpx"
	"github.com/skycode/ojt-management/backend/internal/trainees"
)

type Coordinator struct {
	ID            uuid.UUID `json:"id"`
	Email         string    `json:"email"`
	DisplayName   string    `json:"display_name"`
	AccountStatus string    `json:"account_status"`
	TraineeCount  int       `json:"trainee_count"`
	CreatedAt     time.Time `json:"created_at"`
}

func ListCoordinators(ctx context.Context, pool *db.Pool, p httpx.Page, q string) ([]Coordinator, int64, error) {
	var qArg any
	if q = strings.TrimSpace(q); q != "" {
		qArg = q
	}
	where := `WHERE u.role = 'coordinator'
		AND ($1::text IS NULL OR u.display_name ILIKE '%'||$1||'%' OR u.email ILIKE '%'||$1||'%')`

	var total int64
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM users u `+where, qArg).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := pool.Query(ctx, `
		SELECT u.id, u.email, u.display_name, u.account_status, u.created_at,
		       (SELECT count(*) FROM coordinator_scopes cs WHERE cs.coordinator_user_id = u.id)
		FROM users u `+where+`
		ORDER BY u.display_name LIMIT $2 OFFSET $3`, qArg, p.Limit(), p.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []Coordinator
	for rows.Next() {
		var c Coordinator
		if err := rows.Scan(&c.ID, &c.Email, &c.DisplayName, &c.AccountStatus, &c.CreatedAt, &c.TraineeCount); err != nil {
			return nil, 0, err
		}
		out = append(out, c)
	}
	return out, total, rows.Err()
}

type CreateCoordinatorInput struct {
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

func CreateCoordinator(ctx context.Context, pool *db.Pool, in CreateCoordinatorInput) (*Coordinator, string, error) {
	fields := map[string]string{}
	if strings.TrimSpace(in.Email) == "" || !strings.Contains(in.Email, "@") {
		fields["email"] = "A valid email is required."
	}
	if strings.TrimSpace(in.DisplayName) == "" {
		fields["display_name"] = "Display name is required."
	}
	if len(fields) > 0 {
		return nil, "", httpx.Validation(fields)
	}
	tempPassword, err := trainees.TempPassword()
	if err != nil {
		return nil, "", httpx.Internal(err)
	}
	hash, err := auth.HashPassword(tempPassword)
	if err != nil {
		return nil, "", httpx.Internal(err)
	}
	var c Coordinator
	err = pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, role, display_name)
		VALUES ($1, $2, 'coordinator', $3)
		RETURNING id, email, display_name, account_status, created_at`,
		strings.ToLower(strings.TrimSpace(in.Email)), hash, strings.TrimSpace(in.DisplayName)).
		Scan(&c.ID, &c.Email, &c.DisplayName, &c.AccountStatus, &c.CreatedAt)
	if err != nil {
		if db.IsUniqueViolation(err, "users_email_lower_key") {
			return nil, "", httpx.Validation(map[string]string{"email": "A user with this email already exists."})
		}
		return nil, "", httpx.Internal(err)
	}
	return &c, tempPassword, nil
}

type PatchCoordinatorInput struct {
	DisplayName   *string `json:"display_name"`
	AccountStatus *string `json:"account_status"`
}

func PatchCoordinator(ctx context.Context, pool *db.Pool, id uuid.UUID, in PatchCoordinatorInput) (*Coordinator, error) {
	var c Coordinator
	err := pool.QueryRow(ctx, `
		SELECT id, email, display_name, account_status, created_at FROM users
		WHERE id = $1 AND role = 'coordinator'`, id).
		Scan(&c.ID, &c.Email, &c.DisplayName, &c.AccountStatus, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httpx.ErrNotFound
		}
		return nil, httpx.Internal(err)
	}

	name := c.DisplayName
	if in.DisplayName != nil {
		name = strings.TrimSpace(*in.DisplayName)
	}
	if name == "" {
		return nil, httpx.Validation(map[string]string{"display_name": "Display name is required."})
	}
	status := c.AccountStatus
	if in.AccountStatus != nil {
		if *in.AccountStatus != "active" && *in.AccountStatus != "inactive" && *in.AccountStatus != "locked" {
			return nil, httpx.Validation(map[string]string{"account_status": "Must be active, inactive, or locked."})
		}
		status = *in.AccountStatus
	}

	err = pool.InTx(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx,
			`UPDATE users SET display_name=$2, account_status=$3, updated_at=now() WHERE id=$1`,
			id, name, status); err != nil {
			return err
		}
		if status != "active" && status != c.AccountStatus {
			if _, err := tx.Exec(ctx,
				`UPDATE user_sessions SET revoked_at=now() WHERE user_id=$1 AND revoked_at IS NULL`, id); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, httpx.Internal(err)
	}
	c.DisplayName, c.AccountStatus = name, status
	return &c, nil
}

// SetTraineeScope replaces a coordinator's scope list atomically.
func SetTraineeScope(ctx context.Context, pool *db.Pool, coordinatorID uuid.UUID, traineeIDs []uuid.UUID) error {
	var isCoord bool
	if err := pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND role = 'coordinator')`, coordinatorID).Scan(&isCoord); err != nil {
		return httpx.Internal(err)
	}
	if !isCoord {
		return httpx.ErrNotFound
	}
	for _, id := range traineeIDs {
		var exists bool
		if err := pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM trainee_profiles WHERE id = $1)`, id).Scan(&exists); err != nil {
			return httpx.Internal(err)
		}
		if !exists {
			return httpx.Validation(map[string]string{"trainee_ids": "One or more trainees do not exist."})
		}
	}
	return pool.InTx(ctx, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx,
			`DELETE FROM coordinator_scopes WHERE coordinator_user_id = $1`, coordinatorID); err != nil {
			return err
		}
		for _, tid := range traineeIDs {
			if _, err := tx.Exec(ctx,
				`INSERT INTO coordinator_scopes (coordinator_user_id, trainee_id) VALUES ($1, $2)
				 ON CONFLICT DO NOTHING`, coordinatorID, tid); err != nil {
				return err
			}
		}
		return nil
	})
}
