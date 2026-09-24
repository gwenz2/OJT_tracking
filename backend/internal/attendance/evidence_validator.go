// Evidence image validation: size cap, magic-byte sniffing, real decode, and
// dimension bounds. Filenames from the client are never trusted.
package attendance

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg" // register JPEG decoder
	_ "image/png"  // register PNG decoder

	_ "golang.org/x/image/webp" // register WebP decoder

	"github.com/skycode/ojt-management/backend/internal/platform/httpx"
)

const (
	maxImageDimension = 8000 // px; guards decompression bombs
	minImageDimension = 32   // px; below this it cannot be a real camera photo
)

// ValidatedImage is a decoded, bounds-checked evidence photo.
type ValidatedImage struct {
	Image    image.Image
	MimeType string
	Width    int
	Height   int
	Size     int64
}

var allowedMime = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

// ValidateEvidenceImage checks payload size, content signature, decodability,
// and pixel bounds. data must already be length-capped by the caller (the
// Fiber BodyLimit enforces the hard cap; maxBytes enforces the policy cap).
func ValidateEvidenceImage(data []byte, maxBytes int64) (*ValidatedImage, error) {
	if len(data) == 0 {
		return nil, httpx.Validation(map[string]string{"photo": "A photo is required."})
	}
	if int64(len(data)) > maxBytes {
		return nil, httpx.New(413, "IMAGE_TOO_LARGE",
			fmt.Sprintf("Photo exceeds the %d MB limit.", maxBytes>>20))
	}

	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, httpx.New(415, "INVALID_IMAGE", "File is not a valid image.")
	}
	mime := "image/" + format
	if !allowedMime[mime] {
		return nil, httpx.New(415, "INVALID_IMAGE", "Only JPEG, PNG, or WebP photos are accepted.")
	}

	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w < minImageDimension || h < minImageDimension {
		return nil, httpx.New(415, "INVALID_IMAGE", "Photo is too small to be valid evidence.")
	}
	if w > maxImageDimension || h > maxImageDimension {
		return nil, httpx.New(415, "INVALID_IMAGE", "Photo dimensions exceed limits.")
	}

	return &ValidatedImage{
		Image:    img,
		MimeType: mime,
		Width:    w,
		Height:   h,
		Size:     int64(len(data)),
	}, nil
}
