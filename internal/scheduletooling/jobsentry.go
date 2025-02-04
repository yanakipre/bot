package scheduletooling

import (
	"context"

	"github.com/getsentry/sentry-go"

	"github.com/yanakipre/bot/internal/sentrytooling"
)

func (j *InProcessJob) sentryHub(ctx context.Context, traceID string) (context.Context, func()) {
	ctx = sentrytooling.InitCtx(ctx, "Yanakipre Scheduled Job")
	sCtx := sentrytooling.FromContext(ctx)
	sCtx["Trace ID"] = traceID

	sentry.GetHubFromContext(ctx).Scope().SetTransaction(j.Description())

	span := sentry.StartSpan(ctx, "job_execution")
	return ctx, func() {
		span.Finish()
	}
}
