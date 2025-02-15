package worker

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/yanakipre/bot/internal/logger"
	"github.com/yanakipre/bot/internal/semerr"
	"github.com/yanakipre/bot/internal/sentrytooling"
)

func (j *Middleware) handlePanic(ctx context.Context, name string, start time.Time) func() {
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
			zap.String("job_key", name),
			zap.Duration("runtime", runtime),
			zap.String("op", "complete-job"),
			zap.NamedError("context_error", ctxErr),
		}

		if p == nil {
			return
		}

		logger.Panic(ctx, p, fields...)

		// sentry integration
		sentrytooling.Report(ctx, j.toSemanticErr(semerr.UnwrapPanic(p)))
	}
}
