package ratetooling

import (
	"context"
	"sync"

	"golang.org/x/sync/semaphore"
)

type SemaphoreBasedConfig struct {
	RequestsInFlight int64 `yaml:"requests_in_flight"`
}

func DefaultSemaphoreBasedConfig() SemaphoreBasedConfig {
	return SemaphoreBasedConfig{RequestsInFlight: 1}
}

type SemaphoreBasedLimiter struct {
	mu        sync.Mutex
	container map[string]*semaphore.Weighted
	cfg       SemaphoreBasedConfig
}

func NewSemaphoreBasedLimiter(cfg SemaphoreBasedConfig) *SemaphoreBasedLimiter {
	return &SemaphoreBasedLimiter{container: map[string]*semaphore.Weighted{}, cfg: cfg}
}

const weight = 1

func (p *SemaphoreBasedLimiter) getSemaphore(key string) *semaphore.Weighted {
	var r *semaphore.Weighted
	p.mu.Lock()
	defer p.mu.Unlock()
	value, ok := p.container[key]
	if ok {
		return value
	}

	r = semaphore.NewWeighted(p.cfg.RequestsInFlight)
	p.container[key] = r
	return r
}

func (p *SemaphoreBasedLimiter) Exec(
	ctx context.Context,
	key string,
	exec func() error,
) error {
	r := p.getSemaphore(key)
	if err := r.Acquire(ctx, weight); err != nil {
		return err
	}
	defer r.Release(weight)
	return exec()
}
