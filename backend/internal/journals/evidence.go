// Journal supporting-image storage: private objects, count/size limits,
// owner-only mutation while the journal is editable.
package journals

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/skycode/ojt-management/backend/internal/attendance"
	"github.com/skycode/ojt-management/backend/internal/platform/db"
	"github.com/skycode/ojt-management/backend/internal/platform/httpx"
	"github.com/skycode/ojt-management/backend/internal/platform/storage"
)

// AddEvidence validates + stores a supporting image and links it to the
// journal. Allowed only while the journal is draft/needs_revision.
func AddEvidence(ctx context.Context, pool *db.Pool, store *storage.Store, journalID uuid.UUID, raw []byte, maxBytes int64) (*EvidenceItem, error) {
	img, err := attendance.ValidateEvidenceImage(raw, maxBytes)
	if err != nil {
		return nil, err
	}

	var count int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM journal_evidence WHERE journal_id = $1`, journalID).Scan(&count); err != nil {
		return nil, err
	}
	if count >= maxEvidencePerJournal {
		return nil, httpx.Unprocessable("EVIDENCE_LIMIT",
			fmt.Sprintf("At most %d supporting images per journal.", maxEvidencePerJournal))
	}

	key := fmt.Sprintf("journals/%s.%s", uuid.NewString(),
		map[string]string{"image/jpeg": "jpg", "image/png": "png", "image/webp": "webp"}[img.MimeType])
	sum := sha256.Sum256(raw)
	if err := store.Put(ctx, key, bytes.NewReader(raw), img.Size, img.MimeType); err != nil {
		return nil, httpx.New(503, "STORAGE_UNAVAILABLE", "Storage unavailable.").
			WithInternal(err)
	}

	var item EvidenceItem
	err = pool.InTx(ctx, func(tx pgx.Tx) error {
		if err := lockEditable(ctx, tx, journalID); err != nil {
			return err
		}
		return tx.QueryRow(ctx, `
			INSERT INTO journal_evidence (journal_id, object_key, sha256, mime_type, size_bytes, width_px, height_px)
			VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id, created_at`,
			journalID, key, hex.EncodeToString(sum[:]), img.MimeType, img.Size, img.Width, img.Height).
			Scan(&item.ID, &item.CreatedAt)
	})
	if err != nil {
		_ = store.Delete(ctx, key)
		return nil, err
	}
	item.MimeType, item.SizeBytes, item.Width, item.Height = img.MimeType, img.Size, img.Width, img.Height
	return &item, nil
}

// DeleteEvidence removes a supporting image while the journal is editable.
// The storage object is deleted after the row is removed.
func DeleteEvidence(ctx context.Context, pool *db.Pool, store *storage.Store, journalID, evidenceID uuid.UUID) error {
	var key string
	err := pool.InTx(ctx, func(tx pgx.Tx) error {
		if err := lockEditable(ctx, tx, journalID); err != nil {
			return err
		}
		return tx.QueryRow(ctx, `
			DELETE FROM journal_evidence WHERE id = $1 AND journal_id = $2
			RETURNING object_key`, evidenceID, journalID).Scan(&key)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return httpx.ErrNotFound
	}
	if err != nil {
		return err
	}
	if err := store.Delete(ctx, key); err != nil {
		return httpx.Internal(fmt.Errorf("delete object: %w", err))
	}
	return nil
}

// EvidenceImageKey returns the object key for an evidence item the caller is
// already authorized to see (handler checks ownership/scope first).
func EvidenceImageKey(ctx context.Context, pool *db.Pool, journalID, evidenceID uuid.UUID) (string, error) {
	var key string
	err := pool.QueryRow(ctx,
		`SELECT object_key FROM journal_evidence WHERE id = $1 AND journal_id = $2`,
		evidenceID, journalID).Scan(&key)
	return key, err
}
