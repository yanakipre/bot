// Package rate provides various rate limiters.
package rate

import (
	"sync"
	"time"

	"github.com/yanakipre/bot/internal/encodingtooling"
)

// Window is a simple fixed-window primitive for rate limiting.
// It's not safe to use this from multiple goroutines without additional locking.
type window struct {
	// The number of requests allowed in the window.
	limit uint
	// The duration of the window.
	duration time.Duration
	// The number of requests observed in the window.
	requests uint
	// The time the window started.
	start time.Time
}

// allow returns true and a zero duration if a request is allowed,
// otherwise false and the time the requester has to wait for the request to be possibly allowed.
func (w *window) allow(now time.Time) (bool, time.Duration) {
	// If the window has expired, reset it.
	if now.Sub(w.start) >= w.duration {
		w.start = now
		w.requests = 0
	}

	if w.requests < w.limit {
		w.requests++
		return true, 0
	}

	return false, w.start.Add(w.duration).Sub(now)
}

// wouldAllow is like `allow` but doesn't mutate the internal state of the window.
func (w *window) wouldAllow(now time.Time) (bool, time.Duration) {
	if w.requests < w.limit {
		return true, 0
	}

	return false, w.start.Add(w.duration).Sub(now)
}

type windows []*window

func (ws windows) allow(now time.Time) (bool, time.Duration) {
	for i := range ws {
		ok, wait := ws[i].allow(now)
		if !ok {
			return false, wait
		}
	}

	return true, 0
}

func (ws windows) wouldAllow(now time.Time) (bool, time.Duration) {
	for i := range ws {
		ok, wait := ws[i].wouldAllow(now)
		if !ok {
			return false, wait
		}
	}

	return true, 0
}

// WindowConfig configures a single fixed-window rate limit.
type WindowConfig struct {
	// The number of requests allowed in the window.
	Limit uint `yaml:"limit"    json:"limit"`
	// The duration of the window.
	Duration encodingtooling.Duration `yaml:"duration" json:"duration"`
}

func (wc *WindowConfig) intoWindow() *window {
	return &window{
		limit:    wc.Limit,
		duration: wc.Duration.Duration,
	}
}

// MultiBucketFixedWindowLimiter is a rate limiter that uses multiple fixed windows,
// usually of decreasing RPS, to limit requests.
//
// This is useful if you want to define rate limits like the following:
// - 10 requests/second
// - 300 requests/minute
// - 1500 requests/10 minutes
//
// A request must "fit" in all of those fixed windows to be allowed.
//
// It's safe to use this limiter from multiple goroutines
type MultiBucketFixedWindowLimiter struct {
	mu      sync.Mutex
	windows map[string]windows
	cfg     []WindowConfig
}

// NewMultiBucketFixedWindowLimiter creates a new MultiBucketFixedWindowLimiter from the given config.
func NewMultiBucketFixedWindowLimiter(config []WindowConfig) *MultiBucketFixedWindowLimiter {
	return &MultiBucketFixedWindowLimiter{
		windows: make(map[string]windows),
		cfg:     config,
	}
}

// Allow returns true and zero Duration if the request is allowed,
// or false and the Duration the requester has to wait for the request to be possibly allowed.
func (mb *MultiBucketFixedWindowLimiter) Allow(key string, now time.Time) (bool, time.Duration) {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	ws, ok := mb.windows[key]
	if !ok {
		ws = make(windows, len(mb.cfg))
		for i := range mb.cfg {
			ws[i] = mb.cfg[i].intoWindow()
		}
		mb.windows[key] = ws
	}

	return ws.allow(now)
}

// WouldAllow is like Allow but doesn't mutate the internal state of the limiter.
// It's meant to be called after Allow if there was an override.
func (mb *MultiBucketFixedWindowLimiter) WouldAllow(key string, now time.Time) (bool, time.Duration) {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	ws, ok := mb.windows[key]
	if !ok {
		// The key should be present, but if it's not we'll fail open.
		return true, 0
	}

	return ws.wouldAllow(now)
}

// OverrideWindows overrides the window configuration for a given key.
// The key must be present in the limiter's configuration for it to take effect.
// Only the limit and duration are overridden, the start time and request count are not.
func (mb *MultiBucketFixedWindowLimiter) OverrideWindows(key string, overrides []WindowConfig) {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	ws, ok := mb.windows[key]
	if !ok {
		return
	}

	if len(overrides) > len(ws) {
		// Unexpected configuration - drop the extra windows.
		overrides = overrides[:len(ws)]
	}

	for i := range overrides {
		ws[i].limit = overrides[i].Limit
		ws[i].duration = overrides[i].Duration.Duration
	}
}
