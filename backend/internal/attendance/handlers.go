// HTTP handlers for /api/v1/attendance/*.
package attendance

import (
	"io"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/skycode/ojt-management/backend/internal/auth"
	"github.com/skycode/ojt-management/backend/internal/platform/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// TimeIn: POST /api/v1/attendance/time-in (multipart)
func (h *Handler) TimeIn(c fiber.Ctx) error {
	in, err := parseEvidenceForm(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	res, replayed, err := h.svc.TimeIn(c.Context(), auth.ActorOf(c).ID, *in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	if replayed {
		return httpx.OK(c, res)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": res})
}

// TimeOut: POST /api/v1/attendance/{id}/time-out (multipart)
func (h *Handler) TimeOut(c fiber.Ctx) error {
	attendanceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.Fail(c, httpx.New(404, "ATTENDANCE_NOT_FOUND", "Attendance record not found."))
	}
	in, err := parseEvidenceForm(c)
	if err != nil {
		return httpx.Fail(c, err)
	}
	res, _, err := h.svc.TimeOut(c.Context(), auth.ActorOf(c).ID, attendanceID, *in)
	if err != nil {
		return httpx.Fail(c, err)
	}
	return httpx.OK(c, res) // 200 for both new close and replay per contract
}

// parseEvidenceForm reads the multipart evidence fields. Any client-supplied
// user/site/time/status/watermark fields are ignored — the server owns them.
func parseEvidenceForm(c fiber.Ctx) (*EvidenceInput, error) {
	in := &EvidenceInput{}

	caStr := c.FormValue("client_action_id")
	caID, err := uuid.Parse(caStr)
	if caStr == "" || err != nil {
		return nil, httpx.Validation(map[string]string{
			"client_action_id": "A valid UUID is required.",
		})
	}
	in.ClientActionID = caID

	fh, err := c.FormFile("photo")
	if err != nil {
		return nil, httpx.Validation(map[string]string{"photo": "A photo is required."})
	}
	f, err := fh.Open()
	if err != nil {
		return nil, httpx.New(415, "INVALID_IMAGE", "Could not read the uploaded photo.")
	}
	defer f.Close()
	// Hard cap slightly above the policy cap — the validator returns the
	// proper IMAGE_TOO_LARGE contract error below the transport limit.
	raw, err := io.ReadAll(io.LimitReader(f, 64<<20))
	if err != nil {
		return nil, httpx.New(415, "INVALID_IMAGE", "Could not read the uploaded photo.")
	}
	in.Photo = raw

	if v := c.FormValue("device_captured_at"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			in.DeviceCapturedAt = &t
		}
	}
	if v := c.FormValue("device_timezone_offset_min"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			in.DeviceTZOffsetMin = &n
		}
	}
	in.Latitude = parseFloatField(c, "latitude")
	in.Longitude = parseFloatField(c, "longitude")
	in.GPSAccuracyM = parseFloatField(c, "gps_accuracy_m")
	in.LocationExceptionReason = c.FormValue("location_exception_reason")
	return in, nil
}

func parseFloatField(c fiber.Ctx, name string) *float64 {
	v := c.FormValue(name)
	if v == "" {
		return nil
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return nil // malformed coordinate metadata is treated as absent
	}
	return &f
}
