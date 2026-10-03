package middleware

import (
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// RateLimiter is a fixed-window per-key counter.
//
// A fixed window is chosen over a token bucket because the limit here is a
// handful of submissions per hour: the extra precision of a bucket buys nothing,
// and the simpler structure is easier to reason about when reviewing abuse.
type RateLimiter struct {
	mu       sync.Mutex
	limit    int
	window   time.Duration
	visitors map[string]*window
	stop     chan struct{}
}

type window struct {
	count   int
	resetAt time.Time
}

func NewRateLimiter(limit int, per time.Duration) *RateLimiter {
	rl := &RateLimiter{
		limit:    limit,
		window:   per,
		visitors: make(map[string]*window),
		stop:     make(chan struct{}),
	}
	go rl.prune()
	return rl
}

// Allow reports whether the key may proceed, and when its window resets.
func (rl *RateLimiter) Allow(key string) (bool, time.Time) {
	now := time.Now()

	rl.mu.Lock()
	defer rl.mu.Unlock()

	w, ok := rl.visitors[key]
	if !ok || now.After(w.resetAt) {
		w = &window{count: 0, resetAt: now.Add(rl.window)}
		rl.visitors[key] = w
	}

	if w.count >= rl.limit {
		return false, w.resetAt
	}

	w.count++
	return true, w.resetAt
}

// Close stops the background pruner.
func (rl *RateLimiter) Close() { close(rl.stop) }

// prune drops expired windows so the map does not grow without bound.
func (rl *RateLimiter) prune() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()

	for {
		select {
		case <-rl.stop:
			return
		case now := <-ticker.C:
			rl.mu.Lock()
			for key, w := range rl.visitors {
				if now.After(w.resetAt) {
					delete(rl.visitors, key)
				}
			}
			rl.mu.Unlock()
		}
	}
}

// Limit rejects a request once the caller has exceeded its window.
func Limit(rl *RateLimiter, trustProxy bool, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := ClientIP(r, trustProxy)

			allowed, resetAt := rl.Allow(ip)
			if !allowed {
				retryAfter := max(int(time.Until(resetAt).Seconds())+1, 1)
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))

				log.Warn("rate limit exceeded",
					slog.String("request_id", RequestIDFrom(r.Context())),
					slog.String("ip", ip),
					slog.String("path", r.URL.Path),
				)

				writeJSON(w, http.StatusTooManyRequests,
					`{"ok":false,"message":"Too many messages sent from this address. Please try again later, or email me directly."}`)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
