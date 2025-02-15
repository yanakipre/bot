package worker

import (
	"context"

	"github.com/getsentry/sentry-go"

	"github.com/yanakipre/bot/internal/sentrytooling"
)

func (j *Middleware) sentryHub(ctx context.Context, name string, traceID string) (context.Context, func()) {
	ctx = sentrytooling.InitCtx(ctx, "Yanakipre Scheduled Job")
	sCtx := sentrytooling.FromContext(ctx)
	sCtx["Trace ID"] = traceID

	span := sentry.StartSpan(ctx, "job_execution", sentry.WithTransactionName(name))
	return ctx, func() {
		span.Finish()
	}
}
