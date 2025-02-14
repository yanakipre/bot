package scheduletooling

import (
	"context"

	"github.com/reugn/go-quartz/quartz"
	"go.uber.org/zap"

	"github.com/yanakipre/bot/internal/clouderr"
	"github.com/yanakipre/bot/internal/concurrent"
	"github.com/yanakipre/bot/internal/logger"
	"github.com/yanakipre/bot/internal/scheduletooling/worker"
	"github.com/yanakipre/bot/internal/sentrytooling"
)

type Job interface {
	quartz.Job
	quartz.Trigger
	Key() *quartz.JobKey
	GetConfig() (Config, error)
	SetTrigger(quartz.Trigger)
	Close(ctx context.Context) error
}

type InProcessJob struct {
	close           func(ctx context.Context) error
	runningCount    *concurrent.Value[int64]
	maxRunningCount int64
	trigger         quartz.Trigger
	key             *quartz.JobKey
	name            string
	exec            func(context.Context) error
	getFreshConfig  GetFreshConfig
	metrics         MetricsCollector
}

func (j *InProcessJob) Close(ctx context.Context) error {
	if j.close == nil {
		return nil
	}
	return j.close(ctx)
}

func (j *InProcessJob) SetTrigger(trigger quartz.Trigger) {
	j.trigger = trigger
}

func (j *InProcessJob) GetConfig() (Config, error) {
	cfg, err := j.getFreshConfig()
	if err != nil {
		return Config{}, err
	}
	err = cfg.Validate()
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (j *InProcessJob) Execute(ctx context.Context) error {
	ctx = logger.WithFields(ctx, zap.String("trigger", j.trigger.Description()))
	// attempt to take the latch:
	canRun := false
	j.runningCount.Update(func(runningCount int64) int64 {
		canRun = runningCount < j.maxRunningCount
		if canRun {
			return runningCount + 1
		}
		return runningCount
	})
	if !canRun {
		j.metrics.SkippedJob(j.name)
		msg := "job skipped because max concurrent running jobs reached"
		fields := []zap.Field{
			zap.Int64("max_running", j.maxRunningCount),
		}
		logger.Warn(ctx, msg, fields...)
		return clouderr.WithFields(msg, fields...)
	}
	defer func() {
		p := recover()
		j.runningCount.Update(func(i int64) int64 {
			isPanic := p != nil
			fields := []zap.Field{
				zap.Int64("running_count", i),
				zap.Bool("panic", isPanic),
			}
			if i <= 0 {
				// this should never happen
				err := clouderr.WithFields(
					"update job running count: running count is non-positive",
					fields...,
				)
				logger.Error(ctx, err)
				sentrytooling.Report(ctx, err)
				return 0
			}
			return i - 1
		})
	}()

	return j.exec(ctx)
}

func (j *InProcessJob) Description() string {
	return j.name
}

func (j *InProcessJob) Key() *quartz.JobKey {
	return j.key
}

func (j *InProcessJob) NextFireTime(prev int64) (int64, error) {
	return j.trigger.NextFireTime(prev)
}

type GetFreshConfig func() (Config, error)

func ConstantConfig(config Config) GetFreshConfig {
	return func() (Config, error) {
		return config, nil
	}
}

func NewConcurrentInProcessJobWithCloser(
	execFunc func(context.Context) error,
	closeFunc func(context.Context) error,
	cfg Config,
	getFreshConfig GetFreshConfig,
	metrics MetricsCollector,
	maxRunningCount int64,
) Job {
	mw := worker.NewMiddleware()
	return &InProcessJob{
		close:           closeFunc,
		metrics:         metrics,
		runningCount:    concurrent.NewValue[int64](0),
		maxRunningCount: maxRunningCount,
		trigger:         cfg.GetTriggerValidated(),
		key:             quartz.NewJobKey(cfg.UniqueName),
		name:            cfg.UniqueName,
		exec:            func(ctx context.Context) error { return mw.Execute(ctx, cfg.UniqueName, execFunc) },
		getFreshConfig:  getFreshConfig,
	}
}

func NewInProcessJob(
	exec func(context.Context) error,
	cfg Config,
	getFreshConfig GetFreshConfig,
	metrics MetricsCollector,
) Job {
	return NewConcurrentInProcessJobWithCloser(exec, nil, cfg, getFreshConfig, metrics, 1)
}

func NewConcurrentInProcessJob(
	exec func(context.Context) error,
	cfg Config,
	getFreshConfig GetFreshConfig,
	metrics MetricsCollector,
	maxRunningCount int64,
) Job {
	return NewConcurrentInProcessJobWithCloser(
		exec,
		nil,
		cfg,
		getFreshConfig,
		metrics,
		maxRunningCount,
	)
}
