// Trusted watermark generation. All text is server-built from DB values —
// the client can never supply watermark content (spec FR-020).
package attendance

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"strings"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const watermarkVersion = 1

// WatermarkContext carries the server-trusted values stamped on evidence.
type WatermarkContext struct {
	TraineeName    string
	StudentNumber  string
	Action         string // "Time In" | "Time Out"
	ServerTime     time.Time
	Timezone       string
	SiteName       string
	LocationStatus LocationStatus
}

// BuildWatermarkText returns the two display lines burned into the
// derivative image. Only server values are used.
func BuildWatermarkText(w WatermarkContext) []string {
	local := w.ServerTime
	if loc, err := time.LoadLocation(w.Timezone); err == nil {
		local = w.ServerTime.In(loc)
	}
	return []string{
		fmt.Sprintf("%s (%s) — %s", w.TraineeName, w.StudentNumber, w.Action),
		fmt.Sprintf("%s %s @ %s · GPS: %s",
			local.Format("2006-01-02 15:04:05"), w.Timezone, w.SiteName, w.LocationStatus),
	}
}

// RenderWatermark draws the text block on the bottom of the image and returns
// normalized JPEG bytes (quality 85). The original blob is never modified —
// this output is the reviewable derivative.
func RenderWatermark(src image.Image, lines []string) ([]byte, error) {
	b := src.Bounds()
	dst := image.NewRGBA(b)
	draw.Draw(dst, b, src, b.Min, draw.Src)

	lineH := 16
	pad := 8
	barH := pad*2 + lineH*len(lines)
	barTop := b.Dy() - barH

	// Semi-transparent black strip for legibility over any photo.
	overlay := color.RGBA{0, 0, 0, 160}
	for y := barTop; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			dst.Set(x, y, alphaOver(dst.RGBAAt(x, y), overlay))
		}
	}

	d := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(color.White),
		Face: basicfont.Face7x13,
	}
	for i, line := range lines {
		d.Dot = fixed.P(pad, barTop+pad+i*lineH+12)
		d.DrawString(truncate(line, b.Dx()/8))
	}

	var out bytes.Buffer
	if err := jpeg.Encode(&out, dst, &jpeg.Options{Quality: 85}); err != nil {
		return nil, fmt.Errorf("encode watermark: %w", err)
	}
	return out.Bytes(), nil
}

func alphaOver(bg, fg color.RGBA) color.RGBA {
	a := float64(fg.A) / 255
	return color.RGBA{
		R: uint8(float64(fg.R)*a + float64(bg.R)*(1-a)),
		G: uint8(float64(fg.G)*a + float64(bg.G)*(1-a)),
		B: uint8(float64(fg.B)*a + float64(bg.B)*(1-a)),
		A: 255,
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return strings.TrimSpace(s[:max-1]) + "…"
}
