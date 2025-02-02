package dynamicratelimiter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mitchellh/hashstructure/v2"
	"golang.org/x/time/rate"

	"github.com/yanakipe/bot/internal/logger"
	"github.com/yanakipe/bot/internal/recoverytooling"
	"github.com/yanakipe/bot/internal/resttooling/ratelimiter"
)

type Manager interface {
	Allow(handler string) bool
}

type managerImpl struct {
	// XXX: we can consider using some 'upgradable' RW lock here, like:
	// https://gist.github.com/sancar/d1663e90892cd12c839ae21841b79295
	// but I've seen concerns that it will have benefits if you regularly
	// need to upgrade the lock, which is our case.
	m              sync.Mutex
	limiters       map[string]*rate.Limiter
	config         ratelimiter.RateLimitConfig
	lastConfigHash uint64
	cfgGetter      func() (ratelimiter.RateLimitConfig, error)
}

// Apply rate limit to the given key. If there is no limiter in the internal
// map, new one will be created.
func (rt *managerImpl) Allow(key string) bool {
	rt.m.Lock()
	defer rt.m.Unlock()

	// Consider zero duration as no limit
	if rt.config.Period.Duration == 0 {
		return true
	}

	limiter, ok := rt.limiters[key]
	if !ok {
		// Here we add new key to the limiter, in case of a lot of 404, we can bloat
		// the limiter map. It can probably cause some problems, but so far this seems
		// to be a lesser evil than not rate-limiting at all.
		// TODO: think about shrinking the map periodically.
		limiter = newLimiter(rt.config)
		rt.limiters[key] = limiter
	}

	return limiter.Allow()
}

func NewRateLimitManager(
	ctx context.Context,
	cfgGetter func() (ratelimiter.RateLimitConfig, error),
	cfgRefreshInterval time.Duration,
) (Manager, error) {
	cfg, err := cfgGetter()
	if err != nil {
		return nil, err
	}
	hash, err := hashstructure.Hash(cfg, hashstructure.FormatV2, nil)
	if err != nil {
		return nil, fmt.Errorf("error while hashing rate limiter config: %w", err)
	}

	rlm := &managerImpl{
		limiters:       make(map[string]*rate.Limiter, 1000),
		config:         cfg,
		lastConfigHash: hash,
		cfgGetter:      cfgGetter,
	}

	go recoverytooling.DoUntilSuccess(
		ctx,
		func() error { return rlm.updateLimitersOnConfigChange(ctx, cfgRefreshInterval) },
	)

	return rlm, nil
}

func newLimiter(cfg ratelimiter.RateLimitConfig) *rate.Limiter {
	limiter := rate.NewLimiter(
		rate.Every(
			cfg.Period.Duration/time.Duration(
				// We don't divide by number of workers here, leaving this
				// simple math to the person who will configure it.
				cfg.Requests,
			),
		),
		int(cfg.Burst),
	)

	return limiter
}

func (rt *managerImpl) updateLimitersOnConfigChange(
	ctx context.Context,
	refreshInterval time.Duration,
) error {
	ticker := time.NewTicker(refreshInterval)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			cfg, err := rt.cfgGetter()
			if err != nil {
				return err
			}

			hash, err := hashstructure.Hash(cfg, hashstructure.FormatV2, nil)
			if err != nil {
				return fmt.Errorf("error while hashing rate limiter config: %w", err)
			}

			if hash != rt.lastConfigHash {
				// We don't know, which new keys will come, so just erase all old limiters
				limiters := make(map[string]*rate.Limiter, 1000)
				rt.m.Lock()
				rt.limiters = limiters
				rt.config = cfg
				rt.lastConfigHash = hash
				rt.m.Unlock()

				logger.Info(ctx, "rate limiter config has changed")
			}
		}
	}
}
