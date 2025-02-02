package ratelimiter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mitchellh/hashstructure/v2"
	"golang.org/x/time/rate"

	"github.com/yanakipe/bot/internal/logger"
	"github.com/yanakipe/bot/internal/recoverytooling"
)

const AllHandlers = "all"

type Manager interface {
	Allow(handler string) bool
}

type managerImpl struct {
	cfgGetter      func() (RateLimitersConfig, error)
	limiters       map[string]*rate.Limiter
	m              sync.RWMutex
	lastConfigHash uint64
}

func (rt *managerImpl) Allow(handler string) bool {
	rt.m.RLock()
	defer rt.m.RUnlock()

	limiter, ok := rt.limiters[handler]
	if !ok {
		limiter, ok = rt.limiters[AllHandlers]
		if !ok {
			return true
		}
	}
	return limiter.Allow()
}

func NewRateLimitManager(
	ctx context.Context,
	cfgGetter func() (RateLimitersConfig, error),
) (Manager, error) {
	rateLimitersConfig, err := cfgGetter()
	if err != nil {
		return nil, err
	}
	hash, err := hashstructure.Hash(rateLimitersConfig, hashstructure.FormatV2, nil)
	if err != nil {
		return nil, fmt.Errorf("error while hashing rate limiters config: %w", err)
	}

	limiters := buildLimiters(rateLimitersConfig)

	rlm := &managerImpl{
		cfgGetter:      cfgGetter,
		limiters:       limiters,
		lastConfigHash: hash,
	}

	go recoverytooling.DoUntilSuccess(
		ctx,
		func() error { return rlm.updateLimitersOnConfigChange(ctx) },
	)

	return rlm, nil
}

func buildLimiters(rateLimitersConfig RateLimitersConfig) map[string]*rate.Limiter {
	limiters := make(map[string]*rate.Limiter, len(rateLimitersConfig.ByHandlers))
	for _, cfg := range rateLimitersConfig.ByHandlers {
		limiter := rate.NewLimiter(
			rate.Every(
				cfg.Config.Period.Duration/time.Duration(
					cfg.Config.Requests/rateLimitersConfig.WorkersCount,
				),
			),
			int(cfg.Config.Burst),
		)
		for _, handler := range cfg.Handlers {
			limiters[handler] = limiter
		}
	}

	return limiters
}

func (rt *managerImpl) updateLimitersOnConfigChange(ctx context.Context) error {
	ticker := time.NewTicker(time.Second)
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
				return fmt.Errorf("error while hashing rate limiters config: %w", err)
			}

			if hash != rt.lastConfigHash {
				limiters := buildLimiters(cfg)
				rt.m.Lock()
				rt.limiters = limiters
				rt.lastConfigHash = hash
				rt.m.Unlock()

				logger.Info(ctx, "rate limiters config has changed")
			}
		}
	}
}
