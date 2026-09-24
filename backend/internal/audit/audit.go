// Package audit appends business/security events. Metadata is minimized —
// never passwords, tokens, media bodies, or precise coordinates.
package audit

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

// Executor is satisfied by both *pgxpool.Pool and pgx.Tx.
type Executor interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// Record appends an audit event. It never fails the calling operation — audit
// failure is not a reason to roll back a business transaction.
func Record(ctx context.Context, ex Executor, actor *uuid.UUID, eventType, resourceType string, resourceID *uuid.UUID, requestID string, metadata map[string]any) {
	if metadata == nil {
		metadata = map[string]any{}
	}
	_, _ = ex.Exec(ctx, `
		INSERT INTO audit_events (actor_user_id, event_type, resource_type, resource_id, request_id, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		actor, eventType, resourceType, resourceID, requestID, metadata)
}
