package rate

import (
	"context"
	"sync"

	"golang.org/x/sync/semaphore"
)

// SemaphoreConfig configures the SemaphoreLimiter.
type SemaphoreConfig struct {
	RequestsInFlight int64 `yaml:"requests_in_flight"`
}

// DefaultSemaphoreConfig returns a default configuration for the SemaphoreLimiter.
func DefaultSemaphoreConfig() SemaphoreConfig {
	return SemaphoreConfig{RequestsInFlight: 1}
}

// SemaphoreLimiter is a rate limiter that uses a semaphore to limit the number of concurrent requests, per given key.
type SemaphoreLimiter struct {
	mu        sync.Mutex
	container map[string]*semaphore.Weighted
	cfg       SemaphoreConfig
}

// NewSemaphoreLimiter creates a new SemaphoreLimiter.
func NewSemaphoreLimiter(cfg SemaphoreConfig) *SemaphoreLimiter {
	return &SemaphoreLimiter{
		container: make(map[string]*semaphore.Weighted),
		cfg:       cfg,
	}
}

func (s *SemaphoreLimiter) getSemaphore(key string) *semaphore.Weighted {
	s.mu.Lock()
	defer s.mu.Unlock()

	value, ok := s.container[key]
	if ok {
		return value
	}

	r := semaphore.NewWeighted(s.cfg.RequestsInFlight)
	s.container[key] = r

	return r
}

// Exec executes the given function if the semaphore for the given key allows it.
func (s *SemaphoreLimiter) Exec(
	ctx context.Context,
	key string,
	exec func() error,
) error {
	const weight = 1

	r := s.getSemaphore(key)
	if err := r.Acquire(ctx, weight); err != nil {
		return err
	}
	defer r.Release(weight)

	return exec()
}
