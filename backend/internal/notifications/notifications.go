// In-app notifications. Generation is idempotent via the
// (recipient_user_id, event_key) unique index — repeated job runs or
// transaction replays never duplicate a logical notification.
package notifications

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/skycode/ojt-management/backend/internal/platform/db"
	"github.com/skycode/ojt-management/backend/internal/platform/httpx"
)

// CreateTx inserts a notification inside the caller's transaction.
// Duplicate event keys are ignored (idempotent generation). Returns true when
// a row was actually inserted.
func CreateTx(ctx context.Context, tx pgx.Tx, recipient uuid.UUID, eventKey, ntype, title, body, resourceType string, resourceID *uuid.UUID) (bool, error) {
	res, err := tx.Exec(ctx, `
		INSERT INTO notifications (recipient_user_id, event_key, type, title, body, resource_type, resource_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7) ON CONFLICT DO NOTHING`,
		recipient, eventKey, ntype, title, body, resourceType, resourceID)
	return err == nil && res.RowsAffected() == 1, err
}

// NotifyCompletionIfReachedTx notifies the trainee when the credited minutes
// on the assignment reach its required minutes. Called after any recalculation
// (Time Out, approved correction). The event key is per trainee+assignment so
// it fires exactly once ever.
func NotifyCompletionIfReachedTx(ctx context.Context, tx pgx.Tx, traineeUserID, traineeID, assignmentID uuid.UUID) error {
	var reached bool
	err := tx.QueryRow(ctx, `
		SELECT COALESCE(SUM(s.credited_minutes),0) >= a.required_minutes
		FROM ojt_assignments a
		LEFT JOIN attendance_sessions s
		  ON s.trainee_id = $1 AND s.assignment_id = a.id
		WHERE a.id = $2
		GROUP BY a.required_minutes`, traineeID, assignmentID).Scan(&reached)
	if err != nil || !reached {
		return err
	}
	_, err = CreateTx(ctx, tx, traineeUserID,
		"completion:"+traineeID.String()+":"+assignmentID.String(),
		"completion_milestone", "Required hours completed",
		"Congratulations — you have completed your required OJT hours.",
		"ojt_assignment", &assignmentID)
	return err
}

type Notification struct {
	ID           uuid.UUID  `json:"id"`
	Type         string     `json:"type"`
	Title        string     `json:"title"`
	Body         string     `json:"body"`
	ResourceType *string    `json:"resource_type"`
	ResourceID   *uuid.UUID `json:"resource_id"`
	ReadAt       *time.Time `json:"read_at"`
	CreatedAt    time.Time  `json:"created_at"`
}

// List returns the current user's notifications (optionally unread only).
func List(ctx context.Context, pool *db.Pool, userID uuid.UUID, unreadOnly bool, p httpx.Page) ([]Notification, int64, int64, error) {
	where := `WHERE recipient_user_id = $1 AND ($2::bool = false OR read_at IS NULL)`
	var total, unread int64
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM notifications `+where, userID, unreadOnly).Scan(&total); err != nil {
		return nil, 0, 0, err
	}
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM notifications WHERE recipient_user_id = $1 AND read_at IS NULL`,
		userID).Scan(&unread); err != nil {
		return nil, 0, 0, err
	}
	rows, err := pool.Query(ctx, `
		SELECT id, type, title, body, resource_type, resource_id, read_at, created_at
		FROM notifications `+where+`
		ORDER BY created_at DESC LIMIT $3 OFFSET $4`, userID, unreadOnly, p.Limit(), p.Offset())
	if err != nil {
		return nil, 0, 0, err
	}
	defer rows.Close()
	out := []Notification{}
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.Type, &n.Title, &n.Body,
			&n.ResourceType, &n.ResourceID, &n.ReadAt, &n.CreatedAt); err != nil {
			return nil, 0, 0, err
		}
		out = append(out, n)
	}
	return out, total, unread, rows.Err()
}

// MarkRead marks one notification read for the owner.
func MarkRead(ctx context.Context, pool *db.Pool, userID, id uuid.UUID) error {
	res, err := pool.Exec(ctx, `
		UPDATE notifications SET read_at = now()
		WHERE id = $1 AND recipient_user_id = $2 AND read_at IS NULL`, id, userID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		// Either not found, not owned, or already read — distinguish not found.
		var exists bool
		if err := pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM notifications WHERE id = $1 AND recipient_user_id = $2)`,
			id, userID).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return httpx.ErrNotFound
		}
	}
	return nil
}

// MarkAllRead marks every notification read. Idempotent.
func MarkAllRead(ctx context.Context, pool *db.Pool, userID uuid.UUID) error {
	_, err := pool.Exec(ctx, `
		UPDATE notifications SET read_at = now()
		WHERE recipient_user_id = $1 AND read_at IS NULL`, userID)
	return err
}
