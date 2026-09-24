// Package ratelimit provides a simple in-memory per-key token bucket suitable
// for the single-instance MVP deployment. Per-endpoint policies are applied by
// route (login, attendance writes, correction writes).
package ratelimit

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"golang.org/x/time/rate"

	"github.com/skycode/ojt-management/backend/internal/platform/httpx"
)

type Limiter struct {
	mu      sync.Mutex
	entries map[string]*entry
	limit   rate.Limit
	burst   int
	ttl     time.Duration
}

type entry struct {
	lim      *rate.Limiter
	lastSeen time.Time
}

// New builds a limiter allowing `max` events per `window` with burst = max.
func New(max int, window time.Duration) *Limiter {
	if max <= 0 {
		max = 1
	}
	if window <= 0 {
		window = time.Minute
	}
	return &Limiter{
		entries: make(map[string]*entry),
		limit:   rate.Every(window / time.Duration(max)),
		burst:   max,
		ttl:     window * 4,
	}
}

func (l *Limiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	// lazy cleanup of stale keys
	if len(l.entries) > 10000 {
		cutoff := time.Now().Add(-l.ttl)
		for k, e := range l.entries {
			if e.lastSeen.Before(cutoff) {
				delete(l.entries, k)
			}
		}
	}
	e, ok := l.entries[key]
	if !ok {
		e = &entry{lim: rate.NewLimiter(l.limit, l.burst)}
		l.entries[key] = e
	}
	e.lastSeen = time.Now()
	return e.lim.Allow()
}

// Middleware limits requests per client IP.
func (l *Limiter) Middleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		if !l.allow(c.IP()) {
			return httpx.Fail(c, httpx.ErrRateLimited)
		}
		return c.Next()
	}
}

// KeyedMiddleware limits per arbitrary key (e.g. "login:"+ip,
// "attendance:"+userID).
func (l *Limiter) KeyedMiddleware(keyFn func(fiber.Ctx) string) fiber.Handler {
	return func(c fiber.Ctx) error {
		key := keyFn(c)
		if key == "" {
			key = c.IP()
		}
		if !l.allow(key) {
			return httpx.Fail(c, httpx.ErrRateLimited)
		}
		return c.Next()
	}
}
