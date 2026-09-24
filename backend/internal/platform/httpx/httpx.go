// Package httpx provides the shared HTTP plumbing used by every module:
// request IDs, the stable error envelope from contracts/api-contracts.md,
// and pagination parsing.
package httpx

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

// ---- Error envelope -------------------------------------------------------

// APIError is the stable error contract. Handlers return *APIError (or call
// Fail) and the central ErrorHandler renders:
//
//	{"error": {"code","message","fields","request_id"}}
type APIError struct {
	Status   int               `json:"-"`
	Code     string            `json:"code"`
	Message  string            `json:"message"`
	Fields   map[string]string `json:"fields,omitempty"`
	Internal error             `json:"-"` // logged, never sent to the client
}

func (e *APIError) Error() string {
	if e.Internal != nil {
		return e.Code + ": " + e.Message + ": " + e.Internal.Error()
	}
	return e.Code + ": " + e.Message
}

func New(status int, code, message string) *APIError {
	return &APIError{Status: status, Code: code, Message: message}
}

func (e *APIError) WithFields(fields map[string]string) *APIError {
	e.Fields = fields
	return e
}

func (e *APIError) WithInternal(err error) *APIError {
	e.Internal = err
	return e
}

// Common errors.
var (
	ErrUnauthorized = New(fiber.StatusUnauthorized, "UNAUTHENTICATED", "Authentication is required.")
	ErrForbidden    = New(fiber.StatusForbidden, "FORBIDDEN", "You do not have permission to perform this action.")
	ErrOutOfScope   = New(fiber.StatusForbidden, "OUT_OF_SCOPE", "This trainee is outside your assigned scope.")
	ErrNotFound     = New(fiber.StatusNotFound, "NOT_FOUND", "The requested resource was not found.")
	ErrRateLimited  = New(fiber.StatusTooManyRequests, "RATE_LIMITED", "Too many requests. Please try again shortly.")
	ErrCSRF         = New(fiber.StatusForbidden, "CSRF_REJECTED", "The request could not be verified. Reload and try again.")
)

func Validation(fields map[string]string) *APIError {
	return New(fiber.StatusBadRequest, "VALIDATION_ERROR", "Check the highlighted fields.").WithFields(fields)
}

func ValidationMsg(message string) *APIError {
	return New(fiber.StatusBadRequest, "VALIDATION_ERROR", message)
}

func Conflict(code, message string) *APIError {
	return New(fiber.StatusConflict, code, message)
}

func Unprocessable(code, message string) *APIError {
	return New(fiber.StatusUnprocessableEntity, code, message)
}

func Internal(err error) *APIError {
	return New(fiber.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred.").WithInternal(err)
}

// Fail writes an error response for err. If err is not an *APIError it is
// reported as a generic 500 without leaking internals.
func Fail(c fiber.Ctx, err error) error {
	var ae *APIError
	if !errors.As(err, &ae) {
		ae = Internal(err)
	}
	body := fiber.Map{"error": fiber.Map{
		"code":       ae.Code,
		"message":    ae.Message,
		"request_id": RequestID(c),
	}}
	if len(ae.Fields) > 0 {
		body["error"].(fiber.Map)["fields"] = ae.Fields
	}
	return c.Status(ae.Status).JSON(body)
}

// ErrorHandler is installed as fiber.Config.ErrorHandler.
func ErrorHandler(c fiber.Ctx, err error) error {
	var fe *fiber.Error
	if errors.As(err, &fe) {
		if fe.Code == fiber.StatusRequestEntityTooLarge {
			return Fail(c, New(fiber.StatusRequestEntityTooLarge, "IMAGE_TOO_LARGE", "The uploaded file exceeds the size limit."))
		}
		return Fail(c, New(fe.Code, "HTTP_ERROR", fe.Message))
	}
	return Fail(c, err)
}

// OK writes {"data": v} with an optional meta block.
func OK(c fiber.Ctx, data any) error {
	return c.JSON(fiber.Map{"data": data})
}

func OKMeta(c fiber.Ctx, data any, meta any) error {
	return c.JSON(fiber.Map{"data": data, "meta": meta})
}

// ---- Request ID -----------------------------------------------------------

const requestIDKey = "request_id"

// RequestID middleware assigns/propagates a request ID per request.
func RequestIDMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		id := c.Get("X-Request-ID")
		if id == "" || len(id) > 64 {
			id = "req_" + randHex(12)
		}
		c.Locals(requestIDKey, id)
		c.Set("X-Request-ID", id)
		return c.Next()
	}
}

func RequestID(c fiber.Ctx) string {
	if v, ok := c.Locals(requestIDKey).(string); ok {
		return v
	}
	return ""
}

func randHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// ---- Pagination -----------------------------------------------------------

type Page struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

func (p Page) Offset() int { return (p.Page - 1) * p.PageSize }
func (p Page) Limit() int  { return p.PageSize }

type PageMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}

func NewPageMeta(p Page, total int64) PageMeta {
	tp := total / int64(p.PageSize)
	if total%int64(p.PageSize) != 0 {
		tp++
	}
	return PageMeta{Page: p.Page, PageSize: p.PageSize, Total: total, TotalPages: tp}
}

// ParsePage reads ?page=&page_size= with bounds (1..500, default 25).
func ParsePage(c fiber.Ctx) Page {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	size, _ := strconv.Atoi(c.Query("page_size", "25"))
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 25
	}
	if size > 500 {
		size = 500
	}
	return Page{Page: page, PageSize: size}
}
