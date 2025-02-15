package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/lithammer/shortuuid/v4"
	"go.uber.org/zap"

	"github.com/yanakipre/bot/internal/clouderr"
	"github.com/yanakipre/bot/internal/logger"
)

// Middleware implements all necessary instrumentation:
// logging, metrics, panics handling, etc.
type Middleware struct {
	metrics *wellKnownMetricsCollector
}

func NewMiddleware() *Middleware {
	return &Middleware{
		metrics: NewWellKnownMetricsCollector(),
	}
}

func (j *Middleware) Execute(ctx context.Context, name string, exec func(context.Context) error) error {
	start := time.Now()
	traceID := shortuuid.New()
	// setup the logging so that the error messages get associated
	// with the appropriate log instances
	ctx = logger.WithFields(logger.WithName(ctx, name),
		zap.String("job_key", name),
		zap.String("trace_id", traceID),
		zap.Duration("runtime", time.Since(start)),
	)

	logger.Info(ctx, "job started")

	// create a context before our defer so we can report if the
	// context was canceled externally, given LIFO defer
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ctx, stopSentry := j.sentryHub(ctx, name, traceID)
	defer stopSentry()

	// always execute the following after a job, to prevent process from crashing
	handlePanic := j.handlePanic(ctx, name, start)
	defer handlePanic()

	finished := j.metrics.JobStarted(name)
	err := exec(ctx)
	defer finished(err)
	if err != nil {
		err := clouderr.WrapWithFields(
			fmt.Errorf("job finished with error: %w", err),
			zap.Duration("duration", time.Since(start)),
		)
		logger.Error(ctx, err)
		return err
	}
	logger.Info(ctx, "job finished", zap.Duration("duration", time.Since(start)))
	return nil
}
