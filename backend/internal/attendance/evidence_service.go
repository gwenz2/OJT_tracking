// Evidence storage workflow: validate → store original → render + store
// watermarked derivative → return row metadata. On any mid-flight failure,
// already-stored objects are deleted so the private bucket never orphans.
package attendance

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"

	"github.com/skycode/ojt-management/backend/internal/platform/storage"
)

// StoredEvidence is everything the attendance_evidence INSERT needs that is
// not already known to the transaction.
type StoredEvidence struct {
	OriginalKey     string
	WatermarkedKey  string
	OriginalSHA256  string
	WatermarkSHA256 string
	MimeType        string
	SizeBytes       int64
	Width           int
	Height          int
	WatermarkText   string
}

// StoreEvidence validates the photo, stores the original and the watermarked
// derivative under server-generated keys, and returns insert metadata.
// Cleanup of both objects happens here if any later step fails — callers must
// also call Cleanup on transaction rollback.
func StoreEvidence(ctx context.Context, st *storage.Store, raw []byte, maxBytes int64, wm WatermarkContext) (*StoredEvidence, error) {
	img, err := ValidateEvidenceImage(raw, maxBytes)
	if err != nil {
		return nil, err
	}

	origKey := evidenceKey("original", wm.Action, img.MimeType)
	origSum := sha256.Sum256(raw)
	if err := st.Put(ctx, origKey, bytes.NewReader(raw), img.Size, img.MimeType); err != nil {
		return nil, fmt.Errorf("store original: %w", err)
	}

	lines := BuildWatermarkText(wm)
	wmBytes, err := RenderWatermark(img.Image, lines)
	if err != nil {
		_ = st.Delete(ctx, origKey)
		return nil, fmt.Errorf("render watermark: %w", err)
	}
	wmKey := evidenceKey("watermarked", wm.Action, "image/jpeg")
	wmSum := sha256.Sum256(wmBytes)
	if err := st.Put(ctx, wmKey, bytes.NewReader(wmBytes), int64(len(wmBytes)), "image/jpeg"); err != nil {
		_ = st.Delete(ctx, origKey)
		return nil, fmt.Errorf("store derivative: %w", err)
	}

	return &StoredEvidence{
		OriginalKey:     origKey,
		WatermarkedKey:  wmKey,
		OriginalSHA256:  hex.EncodeToString(origSum[:]),
		WatermarkSHA256: hex.EncodeToString(wmSum[:]),
		MimeType:        img.MimeType,
		SizeBytes:       img.Size,
		Width:           img.Width,
		Height:          img.Height,
		WatermarkText:   joinLines(lines),
	}, nil
}

// Cleanup removes both objects — call it when the DB transaction fails after
// storage succeeded.
func Cleanup(ctx context.Context, st *storage.Store, ev *StoredEvidence) {
	if ev == nil {
		return
	}
	_ = st.Delete(ctx, ev.OriginalKey)
	_ = st.Delete(ctx, ev.WatermarkedKey)
}

// evidenceKey generates the object path. Server-owned: no user filename is
// ever used, preventing path tricks and collisions.
func evidenceKey(kind, action, mime string) string {
	ext := map[string]string{"image/jpeg": "jpg", "image/png": "png", "image/webp": "webp"}[mime]
	return fmt.Sprintf("attendance/%s/%s.%s", kind, uuid.NewString(), ext)
}

func joinLines(lines []string) string {
	out := ""
	for i, l := range lines {
		if i > 0 {
			out += " / "
		}
		out += l
	}
	return out
}
