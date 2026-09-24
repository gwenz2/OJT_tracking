package attendance

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"strings"
	"testing"
	"time"
)

func testJPEG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{uint8(x % 255), uint8(y % 255), 128, 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestValidateEvidenceImage(t *testing.T) {
	valid := testJPEG(t, 640, 480)

	t.Run("valid jpeg", func(t *testing.T) {
		v, err := ValidateEvidenceImage(valid, 6<<20)
		if err != nil {
			t.Fatal(err)
		}
		if v.MimeType != "image/jpeg" || v.Width != 640 || v.Height != 480 {
			t.Fatalf("bad parse: %+v", v)
		}
	})

	t.Run("empty payload", func(t *testing.T) {
		if _, err := ValidateEvidenceImage(nil, 6<<20); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("oversize", func(t *testing.T) {
		if _, err := ValidateEvidenceImage(valid, int64(len(valid)-1)); err == nil {
			t.Fatal("expected size error")
		}
	})

	t.Run("not an image", func(t *testing.T) {
		if _, err := ValidateEvidenceImage([]byte("PK\x03\x04 fake zip"), 6<<20); err == nil {
			t.Fatal("expected decode error")
		}
	})

	t.Run("png accepted", func(t *testing.T) {
		img := image.NewRGBA(image.Rect(0, 0, 100, 100))
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			t.Fatal(err)
		}
		v, err := ValidateEvidenceImage(buf.Bytes(), 6<<20)
		if err != nil || v.MimeType != "image/png" {
			t.Fatalf("png: %v %+v", err, v)
		}
	})

	t.Run("too small rejected", func(t *testing.T) {
		tiny := testJPEG(t, 16, 16)
		if _, err := ValidateEvidenceImage(tiny, 6<<20); err == nil {
			t.Fatal("expected dimension error")
		}
	})

	t.Run("dimension bomb rejected", func(t *testing.T) {
		// Header declares huge dimensions without huge data — PNG max path.
		img := image.NewRGBA(image.Rect(0, 0, maxImageDimension+1, 40))
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			t.Skip("cannot encode huge fixture")
		}
		if _, err := ValidateEvidenceImage(buf.Bytes(), 6<<20); err == nil {
			t.Fatal("expected dimension error")
		}
	})
}

func TestBuildWatermarkText(t *testing.T) {
	lines := BuildWatermarkText(WatermarkContext{
		TraineeName:    "Juan Cruz",
		StudentNumber:  "2024-0001",
		Action:         "Time In",
		ServerTime:     time.Date(2026, 3, 2, 0, 30, 0, 0, time.UTC),
		Timezone:       "Asia/Manila",
		SiteName:       "HQ",
		LocationStatus: LocationVerified,
	})
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %v", lines)
	}
	if !strings.Contains(lines[0], "Juan Cruz") || !strings.Contains(lines[0], "Time In") {
		t.Fatalf("line0 missing identity: %s", lines[0])
	}
	// 00:30 UTC = 08:30 Manila
	if !strings.Contains(lines[1], "08:30:00") || !strings.Contains(lines[1], "Asia/Manila") {
		t.Fatalf("line1 missing local time: %s", lines[1])
	}
	if !strings.Contains(lines[1], "verified") {
		t.Fatalf("line1 missing gps status: %s", lines[1])
	}
}

func TestRenderWatermark(t *testing.T) {
	src, err := jpeg.Decode(bytes.NewReader(testJPEG(t, 400, 300)))
	if err != nil {
		t.Fatal(err)
	}
	out, err := RenderWatermark(src, BuildWatermarkText(WatermarkContext{
		TraineeName: "T", StudentNumber: "1", Action: "Time In",
		ServerTime: time.Now(), Timezone: "Asia/Manila", SiteName: "S",
		LocationStatus: LocationOutsideRadius,
	}))
	if err != nil {
		t.Fatal(err)
	}
	// Output must decode back to a valid JPEG of same dimensions.
	img, format, err := image.Decode(bytes.NewReader(out))
	if err != nil || format != "jpeg" {
		t.Fatalf("derivative not jpeg: %v %s", err, format)
	}
	if img.Bounds().Dx() != 400 || img.Bounds().Dy() != 300 {
		t.Fatal("derivative changed dimensions")
	}
}

func TestEvidenceKey(t *testing.T) {
	k := evidenceKey("original", "time_in", "image/jpeg")
	if !strings.HasPrefix(k, "attendance/original/") || !strings.HasSuffix(k, ".jpg") {
		t.Fatalf("bad key: %s", k)
	}
	// Two keys must differ (uuid segment).
	if evidenceKey("original", "time_in", "image/jpeg") == k {
		t.Fatal("keys collide")
	}
}
