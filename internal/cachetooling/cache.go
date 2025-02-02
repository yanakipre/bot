package cachetooling

import (
	"context"
	"fmt"

	"github.com/lithammer/shortuuid/v4"
	"go.uber.org/zap"

	"github.com/yanakipe/bot/internal/encodingtooling"
	"github.com/yanakipe/bot/internal/logger"
	"github.com/yanakipe/bot/internal/recoverytooling"
	"github.com/yanakipe/bot/internal/scheduletooling"
)

// Cache is a struct that contains all the necessary information to update the cache periodically.
type Cache struct {
	FullUpdateFunc        func(ctx context.Context) error
	IncrementalUpdateFunc func(ctx context.Context) error
	Config                Config
}

type CacheStorage struct {
	caches       []Cache
	jobScheduler *scheduletooling.Scheduler
}

func NewCacheStorage(jobScheduler *scheduletooling.Scheduler) *CacheStorage {
	return &CacheStorage{jobScheduler: jobScheduler}
}

func (cs *CacheStorage) Add(ctx context.Context, c Cache) {
	cs.caches = append(cs.caches, c)
	for _, cacheCfg := range []struct {
		Name        string
		Enabled     bool
		Interval    encodingtooling.Duration
		InitOnStart bool
		UpdateFunc  func(ctx context.Context) error
	}{
		{
			Name:        c.Config.Name + "-full",
			Enabled:     c.Config.FullUpdateEnabled,
			Interval:    c.Config.FullUpdateInterval,
			InitOnStart: c.Config.InitOnStart,
			UpdateFunc:  c.FullUpdateFunc,
		},
		{
			Name:        c.Config.Name + "-incremental",
			Enabled:     c.Config.IncrementalUpdateEnabled,
			Interval:    c.Config.IncrementalUpdateInterval,
			InitOnStart: false,
			UpdateFunc:  c.IncrementalUpdateFunc,
		},
	} {
		ctx := logger.WithFields(ctx, zap.String("cache_name", cacheCfg.Name))
		cacheCfg := cacheCfg

		if !cacheCfg.Enabled {
			logger.Info(ctx, "cache update is disabled")
			continue
		} else if cacheCfg.UpdateFunc == nil {
			logger.Fatal(ctx, "cache update func is not set")
		}

		cfg := scheduletooling.Config{
			UniqueName: cacheCfg.Name,
			Enabled:    cacheCfg.Enabled,
			Interval:   cacheCfg.Interval,
		}
		if err := cs.jobScheduler.Add(ctx, scheduletooling.NewInProcessJob(
			func(ctx context.Context) error {
				if err := cacheCfg.UpdateFunc(ctx); err != nil {
					return fmt.Errorf("could not update cache: %w", err)
				}
				return nil
			},
			cfg,
			func() (scheduletooling.Config, error) {
				return cfg, nil
			},
			&CacheMetricsCollector{},
		)); err != nil {
			logger.Fatal(ctx, "could not add cache update job", zap.Error(err))
		}
	}
}

func (cs *CacheStorage) StartServer(ctx context.Context) {
	for _, c := range cs.caches {
		if c.Config.InitOnStart {
			c := c
			ctx := logger.WithFields(
				ctx,
				zap.String("cache_name", c.Config.Name),
				zap.String("trace_id", shortuuid.New()),
			)

			if c.FullUpdateFunc == nil {
				logger.Error(ctx, "cache full update func is not set")
				continue
			}
			go recoverytooling.DoUntilSuccess(ctx, func() error { return c.FullUpdateFunc(ctx) })
		}
	}
}

func (cs *CacheStorage) ShutdownServer(ctx context.Context) {
	ctx = logger.WithName(ctx, "cache_storage")
	cs.jobScheduler.Wait(ctx)
}
