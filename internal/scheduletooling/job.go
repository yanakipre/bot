package scheduletooling

import (
	"context"
	"time"

	"github.com/lithammer/shortuuid/v4"
	"github.com/reugn/go-quartz/quartz"
	"go.uber.org/zap"

	"github.com/yanakipe/bot/internal/concurrent"
	"github.com/yanakipe/bot/internal/logger"
)

type Job interface {
	quartz.Job
	quartz.Trigger
	GetConfig() (Config, error)
	SetTrigger(quartz.Trigger)
	Close(ctx context.Context) error
}

type InProcessJob struct {
	close           func(ctx context.Context) error
	runningCount    *concurrent.Value[int64]
	maxRunningCount int64
	trigger         quartz.Trigger
	key             *int
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

// loggerContext configures logger in context for background worker.
func loggerContext(ctx context.Context, name string, traceID string) context.Context {
	return logger.WithFields(logger.WithName(ctx, name),
		zap.String("trace_id", traceID),
		zap.String("source", "background_worker"),
		zap.String("worker_type", name),
	)
}

func (j *InProcessJob) Execute(ctx context.Context) {
	start := time.Now()
	traceID := shortuuid.New()
	// setup the logging so that the error messages get associated
	// with the appropriate log instances
	ctx = loggerContext(ctx, j.name, traceID)

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
		logger.Warn(ctx, "job skipped because max concurrent running jobs reached",
			zap.Int("job_id", j.Key()),
			zap.Int64("max_running", j.maxRunningCount),
			zap.Duration("runtime", time.Since(start)),
			zap.String("trigger", j.trigger.Description()),
		)
		return
	}

	logger.Info(ctx, "job started")

	// create a context before our defer so we can report if the
	// context was canceled externally, given LIFO defer
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ctx, stopSentry := j.sentryHub(ctx, traceID)
	defer stopSentry()

	// always execute the following after a job, to prevent process from crashing
	handlePanic := j.handlePanic(ctx, start)
	defer handlePanic()

	finished := j.metrics.JobStarted(j.name)
	defer finished() // record metrics exactly after j.exec returns
	if err := j.exec(ctx); err != nil {
		logger.Error(
			ctx,
			"job finished with error",
			zap.Error(err),
			zap.Duration("duration", time.Since(start)),
		)
		return
	}
	logger.Info(ctx, "job finished", zap.Duration("duration", time.Since(start)))
}

func (j *InProcessJob) Description() string {
	return j.name
}

func (j *InProcessJob) Key() int {
	return *j.key
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
	k := quartz.HashCode(cfg.UniqueName)
	return &InProcessJob{
		close:           closeFunc,
		metrics:         metrics,
		runningCount:    concurrent.NewValue[int64](0),
		maxRunningCount: maxRunningCount,
		trigger:         cfg.GetTriggerValidated(),
		key:             &k,
		name:            cfg.UniqueName,
		exec:            execFunc,
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
