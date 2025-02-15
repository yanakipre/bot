package ratelimiter

import (
	"context"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/yanakipre/bot/internal/clouderr"
	"github.com/yanakipre/bot/internal/logger"
	"github.com/yanakipre/bot/internal/rate"
	"github.com/yanakipre/bot/internal/recoverytooling"
)

// Manager manages rate limiters for different path patterns.
type Manager struct {
	mu sync.RWMutex
	// The value of the appName label in Prometheus metrics reported by the manager.
	appName string
	// Whether request tracking is enabled.
	enabled bool
	// Whether to report or enforce rate limits. Has no effect if enabled is false.
	mode Mode
	// cfgGetter is a function that returns the current configuration, used for live-reloading.
	cfgGetter func() (Config, error)
	// lastConfigHash is used to detect changes in the configuration.
	lastConfigHash uint64
	// Patterns are stored in the order of precedence - most specific patterns are first.
	patterns []*pattern
	// Limiters are stored in the same order as patterns, i.e., the nth limiter corresponds to the nth pattern.
	limiters []*rate.MultiBucketFixedWindowLimiter
	// overrideFetcher fetches overrides for a given key and path pattern
	overrideFetcher func(ctx context.Context, key, pattern string) (bool, []rate.WindowConfig)
}

// NewManager creates a new rate limiter manager.
func NewManager(
	ctx context.Context,
	appName string,
	cfgGetter func() (Config, error),
	overrideFetcher func(ctx context.Context, key, pattern string) (bool, []rate.WindowConfig),
) (*Manager, error) {
	cfg, err := cfgGetter()
	if err != nil {
		return nil, err
	}

	m := &Manager{
		appName:         appName,
		cfgGetter:       cfgGetter,
		overrideFetcher: overrideFetcher,
	}

	if err := m.parseConfig(ctx, cfg); err != nil {
		return nil, err
	}

	go recoverytooling.DoUntilSuccess(ctx, func() error { return m.watchConfig(ctx) })

	return m, nil
}

// Allow returns true and zero Duration if the request is allowed,
// or false and the Duration the requester has to wait for the request to be possibly allowed.
func (m *Manager) Allow(ctx context.Context, method, path, key string) (bool, time.Duration) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.enabled {
		return true, 0
	}

	// Invariant: path[0] == '/'
	path = path[1:]

	// We'll be looping across many patterns, split the path into segments once.
	segments := strings.Split(path, "/")

	for i, p := range m.patterns {
		if !p.match(method, segments) {
			continue
		}

		SeenRequestsTotal.WithLabelValues(m.appName, p.pattern, key).Inc()

		// Invariant: the ith limiter corresponds with the ith pattern.
		limiter := m.limiters[i]

		ok, wait := limiter.Allow(key, time.Now())
		if ok {
			// We know that there are no overlapping requests, because otherwise the config couldn't be parsed.
			// If we matched the pattern to the request, we know that other patterns won't match.
			// So we don't need to check other patterns and their limiters.
			return true, 0
		}

		// Check if there are any overrides for this key and pattern.
		// If there are any, override the windows - it won't reset the request count.
		// Then we can check if the request is allowed, without incrementing the request count.
		hasOverride, overrides := m.overrideFetcher(ctx, key, p.pattern)
		if hasOverride {
			limiter.OverrideWindows(key, overrides)
			ok, wait = limiter.WouldAllow(key, time.Now())
			if ok {
				// Later requests won't log this message, we'll bail out early
				// because the in-memory state of the limiter was already updated.
				logger.Info(
					ctx, "rate limit exceeded, but allowing the request due to an override",
					zap.String("key", key),
					zap.String("pattern", p.pattern),
				)
				return true, 0
			}
		}

		// If we're in reporting mode, then allow the request anyway.
		if m.mode == ModeReporting {
			logger.Info(
				ctx, "rate limit exceeded, but allowing the request in reporting mode",
				zap.String("key", key),
				zap.String("pattern", p.pattern),
				zap.Duration("wait", wait),
			)

			WouldBeRejectedRequestsTotal.WithLabelValues(m.appName, p.pattern, key).Inc()
			return true, 0
		}

		RejectedRequestsTotal.WithLabelValues(m.appName, p.pattern, key).Inc()
		return ok, wait
	}

	return true, 0
}

// parseConfig parses the given configuration and updates the rate limiter manager.
func (m *Manager) parseConfig(ctx context.Context, cfg Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	hash, err := cfg.hash()
	if err != nil {
		return err
	}

	m.enabled = cfg.Enabled
	m.mode = cfg.Mode
	m.lastConfigHash = hash

	// Clearing limiters on config change could mean that we temporarily go over the rate limits,
	// but it's way easier than merging the old and new configurations.
	m.patterns = nil
	m.limiters = nil

	for path, windows := range cfg.Paths {
		p, err := newPattern(path)
		if err != nil {
			return clouderr.WithFields(
				"failed to parse the path pattern",
				zap.Error(err),
				zap.String("path", path),
			)
		}

		// We want to keep patterns sorted by precedence, so we need to find the right place to insert the new pattern.
		// Treat -1 as unset.
		insertionIndex := -1

		// Let's check if this path conflicts with any other path we already know of.
		for idx, p2 := range m.patterns {
			// Two patterns conflict if their relationship is either equivalence (they match the same set of requests)
			// or overlap (they both match some requests, but neither is more specific than the other).
			relation := p.comparePathsAndMethods(p2)
			if relation == equivalent || relation == overlaps {
				return clouderr.WithFields(
					"conflicting path patterns",
					zap.String("new_pattern", p.pattern),
					zap.String("existing_path", p2.pattern),
				)
			}

			// We found the right place to insert the new pattern - it's more specific than the one at this index.
			// Don't break out of the loop though, we need to verify that it doesn't conflict with other patterns.
			if relation == moreSpecific && insertionIndex == -1 {
				insertionIndex = idx
			}
		}

		// If this pattern wasn't more specific than any other pattern, it's the least specific one.
		// We should append it to the end of the list.
		if insertionIndex == -1 {
			insertionIndex = len(m.patterns)
		}

		// Insert the new pattern at the right place, and the limiter at the same index.
		m.patterns = append(m.patterns, nil)
		copy(m.patterns[insertionIndex+1:], m.patterns[insertionIndex:])
		m.patterns[insertionIndex] = p

		m.limiters = append(m.limiters, nil)
		copy(m.limiters[insertionIndex+1:], m.limiters[insertionIndex:])
		m.limiters[insertionIndex] = rate.NewMultiBucketFixedWindowLimiter(windows)
	}

	logger.Info(ctx, "loaded new rate-limiter configuration")

	return nil
}

// watchConfig periodically checks for changes in the config, and updates the manager if necessary.
func (m *Manager) watchConfig(ctx context.Context) error {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			cfg, err := m.cfgGetter()
			if err != nil {
				return err
			}

			hash, err := cfg.hash()
			if err != nil {
				return err
			}

			if hash != m.lastConfigHash {
				return m.parseConfig(ctx, cfg)
			}
		}
	}
}
