package scheduletooling

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/yanakipe/bot/internal/logger"
	"github.com/yanakipe/bot/internal/sentrytooling"
)

func (j *InProcessJob) handlePanic(ctx context.Context, start time.Time) func() {
	return func() {
		// catch the panic to start...
		p := recover()

		// capture the state of the context error *before* we
		// cancel it...
		ctxErr := ctx.Err()

		// record runtime
		runtime := time.Since(start)

		// build logging message, prealloc zap.Field slice for
		// these fields + the one added in the logging
		// statement.
		fields := []zap.Field{
			zap.Int("job_id", j.Key()),
			zap.Duration("runtime", runtime),
			zap.Int64("concurrent_running_count", j.runningCount.Get()),
			zap.String("trigger", j.trigger.Description()),
			zap.String("op", "complete-job"),
			zap.NamedError("context_error", ctxErr),
		}

		j.runningCount.Update(func(i int64) int64 {
			if i <= 0 {
				// this should never happen
				err := fmt.Errorf("running count is non-positive: %d", i)
				logger.Error(ctx, "encountered error", append(fields, zap.Error(err))...)
				sentrytooling.Sentry(ctx, j.toSemanticErr(err))
				return 0
			}
			return i - 1
		})

		if p != nil {
			fields = append(fields,
				zap.Any("panic", p),
				zap.Stack("stack"),
			)
		}

		// sentry integration
		if p != nil {
			err, ok := p.(error)
			if !ok {
				err = fmt.Errorf("panic happened: %v", p)
			}
			sentrytooling.Sentry(ctx, j.toSemanticErr(err))
			logger.Error(ctx, "encountered panic running job", fields...)
		}
	}
}
