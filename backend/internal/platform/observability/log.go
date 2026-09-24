// Package observability provides structured logging with redaction rules for
// secrets, media bodies, and precise location values.
package observability

import (
	"log/slog"
	"os"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/skycode/ojt-management/backend/internal/config"
	"github.com/skycode/ojt-management/backend/internal/platform/httpx"
)

// NewLogger builds the process logger from configuration.
func NewLogger(cfg config.LogConfig) *slog.Logger {
	level := slog.LevelInfo
	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	opts := &slog.HandlerOptions{Level: level, ReplaceAttr: redactAttr}
	var h slog.Handler
	if cfg.Format == "json" {
		h = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		h = slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.New(h)
}

// sensitiveKeys are never logged with their real values.
var sensitiveKeys = map[string]bool{
	"password": true, "password_hash": true, "token": true, "session_token": true,
	"csrf_token": true, "secret": true, "authorization": true, "cookie": true,
	"photo": true, "image": true, "media": true, "object_body": true,
	"latitude": true, "longitude": true, "lat": true, "lng": true,
}

func redactAttr(_ []string, a slog.Attr) slog.Attr {
	k := strings.ToLower(a.Key)
	for part := range strings.SplitSeq(k, ".") {
		if sensitiveKeys[part] {
			a.Value = slog.StringValue("[redacted]")
			return a
		}
	}
	if sensitiveKeys[k] {
		a.Value = slog.StringValue("[redacted]")
	}
	return a
}

// RequestLogger middleware logs one structured line per request.
func RequestLogger(log *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		err := c.Next()
		status := c.Response().StatusCode()
		attrs := []any{
			"request_id", httpx.RequestID(c),
			"method", c.Method(),
			"path", c.Path(),
			"status", status,
			"ip", c.IP(),
		}
		if err != nil {
			attrs = append(attrs, "error", err.Error())
		}
		if status >= 500 {
			log.Error("request", attrs...)
		} else {
			log.Info("request", attrs...)
		}
		return err
	}
}
